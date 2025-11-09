package main

import (
	"context"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"log"
	"os"
	"os/signal"
	"sync"
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

	notifier := newSystemdNotifier()
	quit := make(chan bool)

	sigQuit := make(chan os.Signal, 1)
	signal.Notify(sigQuit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sigQuit
		log.Print("Received signal, initiating graceful shutdown...")
		close(quit)
	}()

	run(mmClient, handler, notifier, quit)
}

func run(mmClient *mmclient.MMClient, handler *handler.Handler, notifier *systemdNotifier, quit chan bool) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		superviseWebSocket(ctx, mmClient, handler, quit)
	}()

	bc := commands.BotCommand{}
	bc.SetPubsub(handler.Pubsub)
	bc.SetCache(handler.Cache)
	bc.ResponseChannel = handler.ResponseChannel
	go commands.Scheduler(bc)

	if notifier != nil {
		notifier.NotifyReady()
		notifier.StartWatchdog(ctx)
	}

	<-quit
	if notifier != nil {
		notifier.NotifyStopping()
	}
	cancel()
	wg.Wait()
	if notifier != nil {
		notifier.Close()
	}
	log.Print("Shutting down...")
	os.Exit(0)
}
