package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/c0dered273/go-adv-metrics/internal/config"
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"github.com/caarlos0/env/v6"
)

func main() {
	var cfg config.Server
	if err := env.Parse(&cfg); err != nil {
		log.Error.Fatal(err)
	}
	cfg.Properties.Repo = storage.GetMemStorageInstance()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	server := &http.Server{Addr: cfg.Address, Handler: handler.Service(cfg)}

	go func() {
		<-shutdown
		shutdownCtx, shutdownCancelCtx := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Error.Fatal("graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := server.Shutdown(shutdownCtx)
		if err != nil {
			log.Error.Fatal(err)
		}

		serverStopCtx()
		shutdownCancelCtx()
	}()

	log.Info.Printf("Metrics server started at %v", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Error.Fatal(err)
	}

	<-serverCtx.Done()
	log.Info.Println("Metrics server shutdown")
}
