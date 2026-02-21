package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"praxis/app/api"
	"praxis/app/lib"
)

func main() {
	cfg, err := lib.LoadConfig()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	appSvc, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("new service: %v", err)
	}
	defer appSvc.Close()

	fiberApp := appSvc.Router()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = fiberApp.ShutdownWithContext(shutdownCtx)
	}()

	log.Printf("listening on %s", cfg.HTTPAddr)
	if err := fiberApp.Listen(cfg.HTTPAddr); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
