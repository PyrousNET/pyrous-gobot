package games

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

type (
	BotGameHelp struct{}

	HelpResponse struct {
		Description string
		Help        string
	}
)

func (gc BotGame) Help(event BotGame) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "dm",
	}

	helpMethods := getHelpMethods()
	helpDocs := compileHelpDocs(helpMethods, event)

	bs := strings.Split(event.body, " ")
	if len(bs) > 0 && bs[0] != "" {
		h, ok := helpDocs[strings.Title(bs[0])]
		if ok {
			response.Message = fmt.Sprintf("```\n%s:\n%s\n```", bs[0], h.Help)
		} else {
			response.Message = fmt.Sprintf("Help for '%s' not found.", bs[0])
		}
	} else {
		mess := "```\nAvailable Games:\n"
		for name, helpDoc := range helpDocs {
			mess += strings.ToLower(name) + " - " + helpDoc.Description + "\n"
		}

		response.Message = fmt.Sprintf("%s```", mess)
	}

	event.ResponseChannel <- response
	return nil
}

func getHelpMethods() []Method {
	methods := []Method{}
	c := BotGameHelp{}
	t := reflect.TypeOf(&c)
	v := reflect.ValueOf(&c)
	for i := 0; i < t.NumMethod(); i++ {
		method := Method{
			typeOf:  t.Method(i),
			valueOf: v.Method(i)}

		methods = append(methods, method)
	}

	return methods
}

func compileHelpDocs(helpMethods []Method, event BotGame) map[string]HelpResponse {
	response := map[string]HelpResponse{}

	for _, m := range helpMethods {
		f := m.valueOf

		in := make([]reflect.Value, 1)
		in[0] = reflect.ValueOf(event)

		var res []reflect.Value
		res = f.Call(in)
		rIface := res[0].Interface()

		response[m.typeOf.Name] = rIface.(HelpResponse)
	}

	return response
}
