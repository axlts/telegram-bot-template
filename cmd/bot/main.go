package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/axlts/telegram-bot-template/internal/config"
	"github.com/axlts/telegram-bot-template/internal/telegram"
)

var (
	configPath = flag.String("config", "", "Path to config file")
)

func main() {
	flag.Parse()

	if *configPath == "" {
		log.Fatal("config file path is required")
	}

	cfg, err := config.Parse(*configPath)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	if err = cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	bot, err := telegram.NewBot(cfg.Bot)
	if err != nil {
		log.Fatalf("bot initialization failed: %v", err)
	}

	ctx := context.Background()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err = bot.Run(ctx); err != nil {
			log.Printf("bot stopped with error: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel()

	if err = bot.Shutdown(ctx); err != nil {
		log.Fatalf("bot shutdown failed: %v", err)
	}

	wg.Wait()
}
