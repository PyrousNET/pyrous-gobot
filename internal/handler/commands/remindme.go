package commands

import (
	"fmt"
	"github.com/olebedev/when"
	"github.com/olebedev/when/rules/en"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"regexp"
	"strings"
	"time"
)

type Reminder struct {
	When time.Time
	Who  string
	What string
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
