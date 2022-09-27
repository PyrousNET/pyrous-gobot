// Copyright (c) 2016 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/handler"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func main() {
	//TODO: Set default env to prod
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	cfg, err := settings.GetConfig(env)
	if err != nil {
		log.Fatalln(err.Error())
	}

	mmClient, err := mmclient.NewMMClient(cfg)
	if err != nil {
		log.Fatalln(err.Error())
	}

	botCache := cache.GetCachingMechanism(cfg.Cache.CONN_STR)

	handler, err := handler.NewHandler(mmClient, botCache)
	if err != nil {
		log.Fatalln(err.Error())
	}

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigQuit
		log.Print("Shutting down...")
		os.Exit(0)

	}()

	run(mmClient, handler)
}

func run(mmClient *mmclient.MMClient, handler *handler.Handler) {
	quit := make(chan bool)

	go func() error {
		for {
			ws, err := mmClient.NewWebSocketClient()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Connected to WS")

			ws.Listen()

			for resp := range ws.EventChannel {
				// We don't want this fella blocking the bot from picking up new events
				go handler.HandleWebSocketResponse(quit, resp)
			}
		}
	}()

	<-quit
	log.Print("Shutting down...")
	os.Exit(2)
}
