package games

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"log"
	"reflect"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"

	"github.com/mattermost/mattermost-server/v6/model"
)

type (
	Games struct {
		availableMethods []Method
		Mm               *mmclient.MMClient
		Settings         *settings.Settings
		Cache            cache.Cache
		GameTigger       string // TODO: This should come from settings
	}

	Method struct {
		typeOf  reflect.Method
		valueOf reflect.Value
	}

	BotGame struct {
		body            string
		sender          string
		target          string
		mm              *mmclient.MMClient
		settings        *settings.Settings
		ReplyChannel    *model.Channel
		method          Method
		ResponseChannel chan comms.Response
		cache           cache.Cache
	}

	Response struct {
		Message string
		Type    string
		Channel string
	}
)

func NewGames(settings *settings.Settings, mm *mmclient.MMClient, cache cache.Cache) *Games {
	games := Games{
		Settings:   settings,
		Mm:         mm,
		Cache:      cache,
		GameTigger: "$", // TODO: This should come from settings
	}

	g := BotGame{}
	t := reflect.TypeOf(&g)
	v := reflect.ValueOf(&g)
	for i := 0; i < t.NumMethod(); i++ {
		method := Method{
			typeOf:  t.Method(i),
			valueOf: v.Method(i)}

		games.availableMethods = append(games.availableMethods, method)
	}

	return &games
}

func (g *Games) NewBotGame(post string, sender string) (BotGame, error) {
	ps := strings.Split(post, " ")

	methodName := strings.Title(strings.TrimLeft(ps[0], g.GameTigger))
	ps = append(ps[:0], ps[1:]...)

	method, err := g.getMethod(methodName)
	if err != nil {
		return BotGame{}, err
	}

	replyChannel := &model.Channel{}
	var rcn string
	if len(ps) > 0 {
		if ps[0] == "in" {
			if len(ps) > 1 {
				rcn = ps[1]
				ps = append(ps[:0], ps[2:]...)

				if rcn != "" {
					c, _ := g.Mm.GetChannel(rcn)
					if c != nil {
						replyChannel = c
					} else {
						log.Default().Println(err)
						return BotGame{}, fmt.Errorf(`The channel "%s" could not be found.`, rcn)
					}

				}
			}
		}
	}

	body := strings.Join(ps[:], " ")

	return BotGame{
		mm:           g.Mm,
		settings:     g.Settings,
		body:         body,
		method:       method,
		ReplyChannel: replyChannel,
		sender:       sender,
		cache:        g.Cache,
	}, nil
}

func (g *Games) CallGame(botGame BotGame) (err error) {
	f := botGame.method.valueOf

	in := make([]reflect.Value, 1)
	in[0] = reflect.ValueOf(botGame)

	var res []reflect.Value
	res = f.Call(in)
	if len(res) > 1 {
		e := res[1].Interface()
		if e != nil {
			err = e.(error)
		}
	}

	return err
}

func (g *Games) getMethod(methodName string) (Method, error) {
	for _, m := range g.availableMethods {
		if m.typeOf.Name == methodName {
			return m, nil
		}
	}

	return Method{}, fmt.Errorf("no such command: %s", methodName)
}
