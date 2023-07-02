package commands

import (
	"encoding/json"
	"fmt"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands/utils"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strconv"
	"strings"
	"time"
)

type Reminder struct {
	Who  string `json:"who"`
	What string `json:"what"`
}

func (h BotCommandHelp) Remindme(request BotCommand) (response HelpResponse) {
	response.Help = `A reminder system with usage of: !remindme <when> {to|about} <what>`
	response.Description = `Bennder will direct message you with a reminder.`

	return response
}

func (bc BotCommand) Remindme(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	pr, when := parseReminder(event.body)
	pr.Who = u.Name
	fmt.Println(pr.Who + " said to remind of " + pr.What + " at " + when.String())
	fmt.Println(pr)

	rmdr, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	timestamp := when.Unix()
	key := "reminders-" + strconv.FormatInt(timestamp, 10)
	bc.pubsub.Set(key, rmdr)
	bc.pubsub.Publish("reminders", key)

	whenStr := when.Format("Mon Jan 2 2006 3:04 PM")

	notify := "Ok. I'll remind you of %s at %s"
	r := comms.Response{
		ReplyChannelId: "",
		Message:        fmt.Sprintf(notify, pr.What, whenStr),
		Type:           "dm",
		UserId:         u.Id,
		Quit:           nil,
	}
	event.ResponseChannel <- r
	return nil
}

func parseReminder(input string) (Reminder, time.Time) {
	var whenStr, what string
	sep := []string{" to ", " about "}
	index := -1
	sepIndex := -1
	for sI, s := range sep {
		sepIndex = sI
		i := strings.Index(input, s)
		if i != -1 {
			index = i
			break
		}
	}

	if index != -1 {
		whenStr = input[:index]
		what = strings.TrimSpace(input[index+len(sep[sepIndex]):])
	}
	w := when.New(nil)
	w.Add(en.All...)

	whenDate, err := w.Parse(whenStr, time.Now())
	if err != nil {
		fmt.Println(err)
	}

	when := whenDate.Time

	return Reminder{
		What: what,
	}, when
}

func Scheduler(bc BotCommand) error {
	// Subscribe to the reminders channel
	pubsub := bc.pubsub.Subscribe("reminders")
	defer pubsub.Close()

	bc.ReloadReminders()

	// Create a channel to receive subscription messages
	channel := pubsub.Channel()

	// Loop forever to process reminders
	for {
		var reminder Reminder
		// Wait for a message on the channel
		msg, ok := <-channel
		if !ok {
			fmt.Println("Message channel closed")
			break
		}

		// Convert the message payload to a timestamp
		timestamp, err := utils.ConvertTimestamp(msg.Payload)
		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			continue
		}

		// Check if the reminder is due
		if time.Now().Unix() >= timestamp {
			// Unmarshal reminder from payload
			reminder, err = bc.GetReminder(msg.Payload)
			if err != nil {
				fmt.Println("Error parsing reminder:", err)
				continue
			}

			// Send the reminder
			err = sendReminder(reminder, bc)
			if err != nil {
				return err
			}

			// Delete the reminder from Redis
			_, err = bc.pubsub.Del(msg.Payload).Result()
			if err != nil {
				fmt.Println("Error deleting reminder:", err)
				continue
			}
		} else {
			key := "reminders-" + strconv.FormatInt(timestamp, 10)
			bc.pubsub.Publish("reminders", key)
		}

		// Add a delay before the next iteration
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (bc *BotCommand) ReloadReminders() error {
	// Load all existing jobs from Redis
	keys, err := bc.cache.GetKeys("reminders-")
	if err != nil {
		return err
	}
	jobs := bc.cache.GetAll(keys)
	if err != nil {
		return err
	}

	for j, _ := range jobs {
		// Process the items retrieved
		fmt.Println("Processing reminder:", j)
		bc.pubsub.Publish("reminders", j)
	}
	return nil
}

func sendReminder(reminder Reminder, bc BotCommand) error {
	u, ok, err := users.GetUser(reminder.Who, bc.cache)
	if err != nil {
		return err
	}

	if ok {
		r := comms.Response{
			ReplyChannelId: "",
			Message:        reminder.What,
			Type:           "dm",
			UserId:         u.Id,
			Quit:           nil,
		}

		bc.ResponseChannel <- r
	} else {
		return fmt.Errorf("scheduler was unable to send %s reminder of %s", reminder.Who, reminder.What)
	}

	return nil
}

func (bc BotCommand) GetReminder(in string) (Reminder, error) {
	var reminder Reminder
	r, err := bc.pubsub.Get(in).Result()
	if err != nil {
		fmt.Println("Error fetching reminder:", err)
		return Reminder{}, err
	}

	// Unmarshal reminder from payload
	err = json.Unmarshal([]byte(r), &reminder)
	if err != nil {
		fmt.Println("Error parsing reminder:", err)
		return Reminder{}, err
	}

	return reminder, nil
}
