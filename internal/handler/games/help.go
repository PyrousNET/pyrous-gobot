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

	args := strings.Fields(event.body)
	if len(args) > 0 {
		query := args[0]
		h, ok := findHelpDoc(helpDocs, query)
		if ok {
			response.Message = fmt.Sprintf("```\n%s:\n%s\n```", query, h.Help)
		} else {
			response.Message = fmt.Sprintf("Help for '%s' not found.", query)
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

func findHelpDoc(helpDocs map[string]HelpResponse, query string) (HelpResponse, bool) {
	for name, doc := range helpDocs {
		if strings.EqualFold(name, query) {
			return doc, true
		}
	}
	return HelpResponse{}, false
}
