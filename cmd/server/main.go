package main

import (
	"context"
	"fmt"
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

//	@Title			Metric Storage API
//	@Description	Сервис хранения метрик.
//	@Version		0.0.1
//  Для сборки сервера с заполнением соответствующих переменных необходимо использовать флаги линковщика
//    go build -ldflags "-X main.buildVersion=v0.0.1 -X 'main.buildDate=$(date +'%Y/%m/%d')' -X 'main.buildCommit=$(git rev-parse HEAD)'"

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	logger := server.NewServerLogger()
	cfg, err := config.NewServerConfig(serverCtx, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("agent: failed to get config")
	}

	httpServer := &http.Server{
		Addr:              cfg.Address,
		ReadHeaderTimeout: 30 * time.Second,
		Handler:           handler.Service(cfg),
	}

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
	err = httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err)
	}

	<-serverCtx.Done()
	logger.Info().Msg("Metrics server shutdown")
}
