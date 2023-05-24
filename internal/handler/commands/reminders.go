package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands/utils"
	"strings"
	"time"
)

import (
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (h BotCommandHelp) Reminders(request BotCommand) (response HelpResponse) {
	response.Help = `List reminders with a usage of: !reminders`
	response.Description = `Bennder will tell you the reminders you have created.`

	return response
}

func (bc BotCommand) Reminders(event BotCommand) error {
	var reminders map[string]Reminder
	reminders = make(map[string]Reminder)
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	// Load all existing jobs from Redis
	keys, err := bc.cache.GetKeys("reminders-")
	if err != nil {
		return err
	}

	for _, k := range keys {
		reminder, err := bc.GetReminder(string(k))

		if err != nil {
			return err
		}
		t, err := utils.ConvertTimestamp(string(k))
		timestamp := time.Unix(t, 0)
		whenStr := timestamp.Format("Mon Jan 2 2006 3:04 PM")
		if err != nil {
			return err
		}

		if reminder.Who == u.Name {
			reminders[whenStr] = reminder
		}
	}

	if len(keys) == 0 {
		r := comms.Response{
			ReplyChannelId: "",
			Message:        "You have no reminders.",
			Type:           "dm",
			UserId:         u.Id,
			Quit:           nil,
		}

		event.ResponseChannel <- r
	} else {
		msg := "| When | What |\n| :------ | :-------|\n"
		for w, r := range reminders {
			msg += fmt.Sprintf("| %s | %s |\n", w, r.What)
		}
		r := comms.Response{
			ReplyChannelId: "",
			Message:        msg,
			Type:           "dm",
			UserId:         u.Id,
			Quit:           nil,
		}

		event.ResponseChannel <- r
	}

	return nil
}
