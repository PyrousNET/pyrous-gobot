package commands

import (
	"encoding/json"
	"fmt"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"regexp"
	"strings"
	"time"
)

type Reminder struct {
	When time.Time `json:"when"`
	Who  string    `json:"who"`
	What string    `json:"what"`
}

func (h BotCommandHelp) Remindme(request BotCommand) (response HelpResponse) {
	response.Help = `A reminder system with usage of: !remindme <when> {to|about} <what>`
	response.Description = `Bennder will direct message you with a reminder.`

	return response
}

func (bc BotCommand) Remindme(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	pr := parseReminder(event.body)
	pr.Who = u.Name
	fmt.Println(pr.Who + " said to remind of " + pr.What + " at " + pr.When.String())
	fmt.Println(pr)

	rmdr, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	bc.pubsub.Publish("reminders", rmdr)

	return nil
}

func parseReminder(input string) Reminder {
	re := regexp.MustCompile(`(.+?)\s+(to|about)\s+(.+)`)
	matches := re.FindStringSubmatch(input)
	whenStr := matches[1]
	what := matches[3]
	w := when.New(nil)
	w.Add(en.All...)

	whenDate, err := w.Parse(whenStr, time.Now())
	if err != nil {
		fmt.Println(err)
	}

	when := whenDate.Time

	return Reminder{
		When: when,
		What: what,
	}
}

func Scheduler(bc BotCommand) error {
	// Subscribe to the reminders channel
	pubsub := bc.pubsub.Subscribe("reminders")
	defer pubsub.Close()

	// Create a channel to receive subscription messages
	channel := pubsub.Channel()

	// Loop forever to process reminders
	for {
		var reminder Reminder
		// Wait for a message on the channel
		msg := <-channel

		// Unmarshal reminder from payload
		err := json.Unmarshal([]byte(msg.Payload), &reminder)
		if err != nil {
			fmt.Println("Error parsing reminder:", err)
			continue
		}

		// Check if the reminder is due
		if time.Now().Unix() >= reminder.When.Unix() {
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
		}

		// Add a delay before the next iteration
		time.Sleep(1 * time.Second)
	}
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
