package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/ruhrcloud/rover/internal/config"
	"github.com/ruhrcloud/rover/internal/tasks"
)

func main() {
	var cfgPath string
	var verbose bool
	flag.StringVar(&cfgPath, "config", "rover.yml", "config file")
	flag.BoolVar(&verbose, "verbose", false, "verbose")
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	err = tasks.Run(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
}

