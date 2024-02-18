package main

import (
	"context"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danthegoodman1/GoAPITemplate/gologger"
	"github.com/danthegoodman1/GoAPITemplate/http_server"
)

var logger = gologger.NewLogger()

func main() {
	if _, err := os.Stat(".env"); err == nil {
		err = godotenv.Load()
		if err != nil {
			logger.Error().Err(err).Msg("error loading .env file, exiting")
			os.Exit(1)
		}
	}

	if len(os.Args) > 1 && os.Args[1] == "coordinator" {
		startCoordinator()
		return
	} else if len(os.Args) > 1 && os.Args[1] == "worker" {
		startWorkerNode()
		return
	}

	logger.Fatal().Msg("must specify coordinator or worker as first argument")
}

func startWorkerNode() {
	logger.Debug().Msg("starting worker node")

}

func startCoordinator() {
	logger.Debug().Msg("starting coordinator api")

	nc, err := nats.Connect(utils.NATS_URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to nats")
	}
	defer nc.Close()

	httpServer := http_server.StartHTTPServer(nc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("failed to shutdown HTTP server")
	} else {
		logger.Info().Msg("successfully shutdown HTTP server")
	}
}
