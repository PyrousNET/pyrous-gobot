package commands

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"

	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type (
	Commands struct {
		availableMethods []Method
		Mm               *mmclient.MMClient
		Settings         *settings.Settings
		Cache            cache.Cache
	}

	Method struct {
		typeOf  reflect.Method
		valueOf reflect.Value
	}

	BotCommand struct {
		body         string
		sender       string
		target       string
		mm           *mmclient.MMClient
		settings     *settings.Settings
		ReplyChannel *model.Channel
		method       Method
		cache        cache.Cache
	}

	Response struct {
		Channel  string
		Delay    time.Duration // Delay for 2nd message to be sent
		Message  string
		Message2 string
		Type     string
	}
)

func NewCommands(settings *settings.Settings, mm *mmclient.MMClient, cache cache.Cache) *Commands {
	commands := Commands{
		Settings: settings,
		Mm:       mm,
		Cache:    cache,
	}

	c := BotCommand{}
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

	methodName := strings.Title(strings.TrimLeft(ps[0], c.Settings.GetCommandTrigger()))
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
		method:       method,
		ReplyChannel: replyChannel,
		sender:       sender,
		cache:        c.Cache,
	}, nil
}

func (c *Commands) CallCommand(botCommand BotCommand) (response Response, err error) {
	f := botCommand.method.valueOf

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(botCommand)

	var res []reflect.Value
	res = f.Call(in)
	rIface := res[0].Interface()
	if len(res) > 1 {
		e := res[1].Interface()
		if e != nil {
			err = e.(error)
		}
	}

	return rIface.(Response), err
}

func (c *Commands) getMethod(methodName string) (Method, error) {
	for _, m := range c.availableMethods {
		if m.typeOf.Name == methodName {
			return m, nil
		}
	}

	return Method{}, fmt.Errorf("no such command: %s", methodName)
}
