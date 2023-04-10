package main

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/handler"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func main() {
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

	// Keep the bot from going inactive
	go func() {
		for {
			err := mmClient.KeepBotActive()
			if err != nil {
				log.Printf("error keeping the bot active: %v", err)
			}
			time.Sleep(290 * time.Second)
		}
	}()

	botCache := cache.GetCachingMechanism(cfg.Server.CACHE_URI)
	botPubsub := pubsub.GetPubsub(cfg.Server.CACHE_URI)

	handler, err := handler.NewHandler(mmClient, botCache, botPubsub)
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
