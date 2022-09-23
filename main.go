// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package main

import (
	"fmt"
	"log"

	"github.com/pyrousnet/mattermost-golang-bot/internal/handler"
	"github.com/pyrousnet/mattermost-golang-bot/internal/mmclient"
)

func main() {
	mm, err := mmclient.NewMMClient()
	if err != nil {
		log.Fatalln(err.Error())
	}

	handler, err := handler.NewHandler(mm)
	if err != nil {
		log.Fatalln(err.Error())
	}

	// Lets start listening to some channels via the websocket!
	for {
		ws, err := mm.NewWebSocketClient()
		if err != nil {
			log.Fatalf(err.Error())
		}

		fmt.Println("Connected to WS")

		ws.Listen()

		for resp := range ws.EventChannel {
			// We don't want this fella blocking the bot from picking up new events
			go handler.HandleWebSocketResponse(resp)
		}
	}
}
