package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charlieegan3/demo-live-policy-update/pkg/server"
	"github.com/charlieegan3/demo-live-policy-update/pkg/server/config"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalln("set config file first arg")
	}

	cfgBS, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}

	cfg, err := config.ParseConfig(cfgBS)
	if err != nil {
		log.Fatalf("failed to parse config: %s", err)
	}

	svr, err := server.NewServer(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	log.Printf("starting server on http://localhost:%d", cfg.Port)

	err = svr.Start(ctx)
	if err != nil {
		log.Fatalf("failed to start server: %s", err)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCtxCancel()

	err = svr.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("failed to stop server: %s", err)
	}

	cancel()
}
