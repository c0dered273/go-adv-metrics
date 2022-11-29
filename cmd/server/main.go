package main

import (
	"context"
	"github.com/c0dered273/go-adv-metrics/internal/handler"
	"github.com/c0dered273/go-adv-metrics/internal/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	serverAddr = "localhost:8080"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	server := &http.Server{Addr: serverAddr, Handler: handler.Service()}

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
