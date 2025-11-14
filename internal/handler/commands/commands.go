package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"

	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"

	"github.com/mattermost/mattermost/server/public/model"
)

type (
	Commands struct {
		availableMethods []Method
		Mm               *mmclient.MMClient
		Settings         *settings.Settings
		Cache            cache.Cache
		Pubsub           pubsub.Pubsub
	}

	Method struct {
		typeOf  reflect.Method
		valueOf reflect.Value
	}

	BotCommand struct {
		body            string
		sender          string
		target          string
		mm              *mmclient.MMClient
		settings        *settings.Settings
		ReplyChannel    *model.Channel
		ResponseChannel chan comms.Response
		method          Method
		cache           cache.Cache
		pubsub          pubsub.Pubsub
		Quit            chan bool
	}

	Response struct {
		Channel  string
		Delay    time.Duration // Delay for 2nd message to be sent
		Message  string
		Message2 string
		Type     string
	}
)

func NewCommands(settings *settings.Settings, mm *mmclient.MMClient, cache cache.Cache, pubsub pubsub.Pubsub) *Commands {
	commands := Commands{
		Settings: settings,
		Mm:       mm,
		Cache:    cache,
		Pubsub:   pubsub,
	}

	c := BotCommand{
		cache:  cache,
		pubsub: pubsub,
	}
	t := reflect.TypeOf(&c)
	v := reflect.ValueOf(&c)
	for i := 0; i < t.NumMethod(); i++ {
		method := Method{
			typeOf:  t.Method(i),
			valueOf: v.Method(i)}

		commands.availableMethods = append(commands.availableMethods, method)
	}

	return &commands
}

func (c *Commands) NewBotCommand(post string, sender string) (BotCommand, error) {
	ps := strings.Split(post, " ")

	if len(ps) == 0 || ps[0] == "" {
		return BotCommand{}, fmt.Errorf("no command provided")
	}

	commandToken := strings.TrimSpace(ps[0])
	trigger := c.Settings.GetCommandTrigger()
	if trigger != "" {
		if strings.HasPrefix(commandToken, trigger) {
			commandToken = strings.TrimPrefix(commandToken, trigger)
		} else {
			commandToken = strings.TrimLeft(commandToken, trigger)
		}
	}

	commandToken = strings.TrimSpace(commandToken)
	methodName := strings.Title(commandToken)
	ps = append(ps[:0], ps[1:]...)

	method, err := c.getMethod(methodName)
	if err != nil {
		return BotCommand{}, err
	}

	replyChannel := &model.Channel{}
	var rcn string
	if len(ps) > 0 {
		if ps[0] == "in" {
			if len(ps) > 1 {
				rcn = ps[1]
				ps = append(ps[:0], ps[2:]...)

				if rcn != "" {
					c, _ := c.Mm.GetChannel(rcn)
					if c != nil {
						replyChannel = c
					} else {
						log.Default().Println(err)
						return BotCommand{}, fmt.Errorf(`The channel "%s" could not be found.`, rcn)
					}

				}
			}
		}
	}

	body := strings.Join(ps[:], " ")

	return BotCommand{
		mm:           c.Mm,
		settings:     c.Settings,
		body:         body,
		target:       commandToken,
		method:       method,
		ReplyChannel: replyChannel,
		sender:       sender,
		cache:        c.Cache,
		pubsub:       c.Pubsub,
	}, nil
}

func (c *Commands) CallCommand(botCommand BotCommand) error {
	var err error
	f := botCommand.method.valueOf

	if botCommand.method.typeOf.Type == nil {
		return fmt.Errorf("Man! What are you talking about? You need `!help`")
	}

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(botCommand)

	var res []reflect.Value
	res = f.Call(in)
	if len(res) > 0 {
		e := res[0].Interface()
		if e != nil {
			err = e.(error)
		}
	}

	return err
}

func (c *Commands) getMethod(methodName string) (Method, error) {
	for _, m := range c.availableMethods {
		if m.typeOf.Name == methodName {
			return m, nil
		}
	}

	return Method{}, fmt.Errorf("no such command: %s", methodName)
}

func (bc *BotCommand) SetCache(c cache.Cache) {
	bc.cache = c
}

func (bc *BotCommand) SetPubsub(ps pubsub.Pubsub) {
	bc.pubsub = ps
}

var diceCommandPattern = regexp.MustCompile(`(?i)^(\d*)d(\d+)$`)

const maxDice = 50

func parseRollSpec(body string) (int, int, string, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return 2, 6, "", nil
	}

	parts := strings.Fields(body)
	if len(parts) == 0 {
		return 2, 6, "", nil
	}

	spec := strings.ToLower(parts[0])
	rest := strings.Join(parts[1:], " ")
	matches := diceCommandPattern.FindStringSubmatch(spec)
	if len(matches) == 0 {
		return 2, 6, body, nil
	}

	count := 1
	if matches[1] != "" {
		var err error
		count, err = strconv.Atoi(matches[1])
		if err != nil || count <= 0 {
			return 0, 0, "", fmt.Errorf("invalid dice count '%s'", matches[1])
		}
	}

	sides, err := strconv.Atoi(matches[2])
	if err != nil || sides <= 0 {
		return 0, 0, "", fmt.Errorf("invalid dice sides '%s'", matches[2])
	}

	if count > maxDice {
		return 0, 0, "", fmt.Errorf("that's too many dice! Please roll %d or fewer.", maxDice)
	}

	return count, sides, strings.TrimSpace(rest), nil
}
