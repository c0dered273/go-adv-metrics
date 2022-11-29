package main

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"github.com/c0dered273/go-adv-metrics/internal/storage"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serverAddr = "localhost"
	serverPort = ":8080"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    serverAddr + serverPort,
		Handler: handler.NewMetricHandler(storage.NewMemStorage()),
	}

	go func() {
		log.Info.Printf("Metrics server started at %v", serverAddr+serverPort)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error.Fatalf("HTTP server ListenAndServe Error: %v", err)
		}
	}()

	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error.Printf("HTTP Server Shutdown Error: %v", err)
	}
	log.Info.Println("Metrics server shutdown")
}
