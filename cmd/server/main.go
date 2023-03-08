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
	"github.com/c0dered273/go-adv-metrics/internal/log/server"
	"github.com/rs/zerolog/log"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	logger := server.NewServerLogger()
	cfg := config.NewServerConfig(serverCtx, logger, config.GetServerConfig())
	httpServer := &http.Server{Addr: cfg.Address, Handler: handler.Service(cfg)}

	go func() {
		<-shutdown
		shutdownCtx, shutdownCancelCtx := context.WithTimeout(serverCtx, 30*time.Second)

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("server: graceful shutdown timed out.. forcing exit.")
			}
		}()

		err := httpServer.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Err(err)
		}

		serverStopCtx()
		shutdownCancelCtx()
	}()

	logger.Info().Msgf("Metrics server started at %v", httpServer.Addr)
	err := httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err)
	}

	<-serverCtx.Done()
	logger.Info().Msg("Metrics server shutdown")
}
