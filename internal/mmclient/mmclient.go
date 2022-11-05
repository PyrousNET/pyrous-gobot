package mmclient

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

type MMClient struct {
	Client           *model.Client4
	WebSocketClient  *model.WebSocketClient
	BotUser          *model.User
	BotTeam          *model.Team
	DebuggingChannel *model.Channel
	Server           Server
	SettingsUrl      string
	cfg              *settings.Config
}

type Server struct {
	HOST      string `yaml:"host"`
	PROTOCOL  string `yaml:"protocol"`
	PORT      string `yaml:"port"`
	CACHE_URI string `yaml:"cache_uri"`
}

// Documentation for the Go driver can be found
// at https://godoc.org/github.com/mattermost/platform/model#Client
func NewMMClient(cfg *settings.Config) (client *MMClient, err error) {
	client = &MMClient{}

	client.cfg = cfg
	client.Server = client.cfg.Server
	client.SettingsUrl = client.cfg.Bot.SETTINGS_URL
	conn := client.Server.PROTOCOL + client.Server.HOST
	client.Client = model.NewAPIv4Client(conn)

	// Lets test to see if the mattermost server is up and running
	client.PingServer()

	// lets attempt to login to the Mattermost server as the bot user
	// This will set the token required for all future calls
	// You can get this token with client.AuthToken
	user, err := client.LoginAsUser()
	if err != nil {
		return client, err
	}

	client.BotUser = user

	// If the bot user doesn't have the correct information lets update his profile
	err = client.UpdateUserIfNeeded()
	if err != nil {
		return client, err
	}

	// Lets find our bot team
	team, err := client.GetTeam()
	if err != nil {
		return client, err
	}

	client.BotTeam = team

	// This is an important step.  Lets make sure we use the botTeam
	// for all future web service requests that require a team.
	//client.SetTeamId(botTeam.Id)

	// Lets create a bot channel for logging debug messages into
	err = client.CreateDebuggingChannelIfNeeded()
	if err != nil {
		return client, err
	}

	client.SendMsgToDebuggingChannel("_"+client.cfg.Bot.SAMPLE_NAME+" has **started** running_", "")

	return client, err
}

func (c *MMClient) SetupGracefulShutdown() {
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt)
	go func() {
		for _ = range channel {
			if c.WebSocketClient != nil {
				c.WebSocketClient.Close()
			}

			err := c.SendMsgToDebuggingChannel("_"+c.cfg.Bot.SAMPLE_NAME+" has **stopped** running_", "")
			if err != nil {
				log.Fatalln(err.Error())
			}

			os.Exit(0)
		}
	}()
}

func (c *MMClient) SendMsgToDebuggingChannel(msg string, replyToId string) error {
	post := &model.Post{}
	post.ChannelId = c.DebuggingChannel.Id
	post.Message = msg

	post.RootId = replyToId

	if _, _, err := c.Client.CreatePost(post); err != nil {
		return fmt.Errorf("We failed to send a message to the logging channel: %+v", err)
	}

	return nil
}

func (c *MMClient) PingServer() {
	if props, _, err := c.Client.GetOldClientConfig(""); err != nil {
		e := fmt.Errorf("There was a problem pinging the Mattermost server.  Are you sure it's running? Error: %+v", err)
		log.Fatalln(e.Error())
	} else {
		log.Println("Server detected and is running version " + props["Version"])
	}
}

func (c *MMClient) LoginAsUser() (*model.User, error) {
	var err error

	user, _, err := c.Client.Login(
		c.cfg.Bot.USER_EMAIL,
		c.cfg.Bot.USER_PASSWORD)

	if err != nil {
		err = fmt.Errorf("There was a problem logging into the Mattermost server. Error: %+v", err)
		return user, err
	}
	return user, err
}

func (c *MMClient) UpdateUserIfNeeded() error {
	if c.BotUser.FirstName != c.cfg.Bot.USER_FIRST || c.BotUser.LastName != c.cfg.Bot.USER_LAST || c.BotUser.Username != c.cfg.Bot.USERNAME {
		c.BotUser.FirstName = c.cfg.Bot.USER_FIRST
		c.BotUser.LastName = c.cfg.Bot.USER_LAST
		c.BotUser.Username = c.cfg.Bot.USERNAME

		user, _, err := c.Client.UpdateUser(c.BotUser)
		if err != nil {
			return fmt.Errorf("Failed to update bot user. Error: %+v", err)
		}
		c.BotUser = user
		log.Println("Updated bot account settings")
	}

	return nil
}

func (c *MMClient) GetTeam() (*model.Team, error) {
	var err error

	team, _, err := c.Client.GetTeamByName(c.cfg.Bot.TEAM_NAME, "")
	if err != nil {
		err = fmt.Errorf("Failed to find team. Error: %+v", err)
	}

	return team, err
}

func (c *MMClient) CreateDebuggingChannelIfNeeded() error {
	log.Println("Attempting to open channel " + c.cfg.Bot.LOG_NAME)

	rchannel, _, err := c.Client.GetChannelByName(c.cfg.Bot.LOG_NAME, c.BotTeam.Id, "")
	if err == nil {
		c.DebuggingChannel = rchannel
		return nil
	}

	// Looks like we need to create the logging channel
	channel := &model.Channel{}
	channel.Name = c.cfg.Bot.LOG_NAME
	channel.DisplayName = "Debugging For Sample Bot"
	channel.Purpose = "This is used as a test channel for logging bot debug messages"
	channel.Type = model.ChannelTypeOpen
	channel.TeamId = c.BotTeam.Id

	rchannel, _, err = c.Client.CreateChannel(channel)
	if err != nil {
		return fmt.Errorf("Failed to create debug channel. Error: %+v", err)
	}

	c.DebuggingChannel = rchannel

	return nil
}

// This function came from the original sample code. It sucks. Use GetChannelByName instead.
func (c *MMClient) GetChannel(name string) (*model.Channel, error) {
	channel, _, err := c.Client.GetChannelByName(name, c.BotTeam.Id, "")
	return channel, err
}

// This function returns a proper error so you can know what the heck is going on
func (c *MMClient) GetChannelByName(name string) (*model.Channel, error) {
	ch, _, err := c.Client.GetChannelByName(name, c.BotTeam.Id, "")
	if err != nil {
		return nil, err
	}

	return ch, nil
}

func (c *MMClient) SendCmdToChannel(cmd string, channelId string, prePost *model.Post) error {
	_, _, err := c.Client.ExecuteCommand(channelId, cmd)
	if err != nil {
		return fmt.Errorf("Failed to send a message to %s. Error: %+v", channelId, err)
	}

	return nil
}

func (c *MMClient) SendMsgToChannel(msg string, channelId string, prePost *model.Post) error {
	post := &model.Post{}
	post.ChannelId = channelId
	post.Message = msg

	if prePost.ReplyCount == 0 {
		post.RootId = prePost.Id
	} else {
		post.RootId = prePost.RootId
	}

	_, _, err := c.Client.CreatePost(post)
	if err != nil {
		return fmt.Errorf("Failed to send a message to %s. Error: %+v", channelId, err)
	}

	return nil
}

func (c *MMClient) NewWebSocketClient() (*model.WebSocketClient, error) {
	var err error
	uri := fmt.Sprintf("ws://%s:%s", c.Server.HOST, c.Server.PORT)

	ws, appErr := model.NewWebSocketClient4(uri, c.Client.AuthToken)
	if appErr != nil {
		err = fmt.Errorf("%+v", appErr)
	}

	return ws, err
}

func (b *MMClient) KeepBotActive() error {
	status, _, err := b.Client.GetUserStatus(b.BotUser.Id, "")
	if err != nil {
		return err
	}
	status.Status = "online"

	_, _, err = b.Client.UpdateUserStatus(b.BotUser.Id, status)

	return err
}

func (b *MMClient) PostFromJson(data io.Reader) *model.Post {
	var o *model.Post
	json.NewDecoder(data).Decode(&o)
	return o
}
