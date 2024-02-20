package main

import (
	"context"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/scheduling"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"go.uber.org/atomic"
	"os"
	"os/signal"
	"sync"
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

	logger.Debug().Msgf("connecting to nats at %s", utils.NATS_URL)
	nc, err := nats.Connect(utils.NATS_URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect to nats")
	}
	defer nc.Close()

	if len(os.Args) > 1 && os.Args[1] == "coordinator" {
		startCoordinator(nc)
		return
	} else if len(os.Args) > 1 && os.Args[1] == "worker" {
		startWorkerNode(nc)
		return
	}

	logger.Fatal().Msg("must specify coordinator or worker as first argument")
}

func startWorkerNode(nc *nats.Conn) {
	logger.Debug().Msgf("starting worker node (%s)", utils.WORKER_ID)

	availableSlots := atomic.NewInt64(utils.SLOTS)

	// We need a sync map to track reservations
	reservations := sync.Map{}

	// Scheduling loop
	_, err := nc.Subscribe("scheduling.request.*", func(msg *nats.Msg) {
		logger.Debug().Msgf("Worker %s got scheduling request, reserving resources", utils.WORKER_ID)
		var request scheduling.ScheduleRequest
		utils.JSONMustUnmarshal(msg.Data, &request)

		// Check whether the region matches (if provided)
		if request.Requirements.Region != "" && request.Requirements.Region != utils.REGION {
			logger.Debug().Msgf(
				"worker %s cannot fulfill request, different region",
				utils.WORKER_ID,
			)
		}

		// Check whether we have enough available slots
		if request.Requirements.Slots > availableSlots.Load() {
			logger.Debug().Msgf(
				"worker %s cannot fulfill request, not enough slots",
				utils.WORKER_ID,
			)
		}

		// Reserve the slots
		// Note: would need better handling to protect against going negative in prod
		availableSlots.Sub(request.Requirements.Slots)
		reservations.Store(request.RequestID, request.Requirements.Slots)

		err := msg.Respond(utils.JSONMustMarshal(scheduling.ScheduleResponse{
			WorkerID: utils.WORKER_ID,
		}))
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to respond to resource request message")
		}

	})
	if err != nil {
		logger.Fatal().Err(err).Msg("error subscribing to scheduling.request.*")
	}

	// Release loop
	_, err = nc.Subscribe("scheduling.release", func(msg *nats.Msg) {
		var payload scheduling.ReleaseResourcesMessage
		utils.JSONMustUnmarshal(msg.Data, &payload)
		if payload.ExemptWorker == utils.WORKER_ID {
			// We are exempt from this
			return
		}

		slots, found := reservations.LoadAndDelete(payload.RequestID)
		if !found {
			logger.Fatal().Msgf(
				"did not find reservation for request %s, crashing to reset!",
				payload.RequestID,
			)
		}
		availableSlots.Add(slots.(int64))

		logger.Debug().Msgf("Worker %s releasing resources", utils.WORKER_ID)
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("error subscribing to scheduling.release")
	}

	_, err = nc.Subscribe(fmt.Sprintf("scheduling.reserve_task.%s", utils.WORKER_ID), func(msg *nats.Msg) {
		// Listen for our own reservations
		var reservation scheduling.ReserveRequest
		utils.JSONMustUnmarshal(msg.Data, &reservation)
		logger.Debug().Msgf("Got reservation on worker node %s with payload %+v", utils.WORKER_ID, reservation)

		if sleepSec, ok := reservation.Payload["SleepSec"].(float64); ok {
			logger.Debug().Msgf(
				"worker %s sleeping for %d seconds",
				utils.WORKER_ID,
				sleepSec,
			)
			time.Sleep(time.Second * time.Duration(sleepSec))
		}

		err = msg.Respond(utils.JSONMustMarshal(scheduling.ReserveResponse{
			Error: nil,
			Payload: map[string]any{ // float64 because of JSON
				"Num": reservation.Payload["Num"].(float64) + 1,
			},
		}))
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to respond to reservation request")
		}

		// we are done, we can release resources
		slots, found := reservations.LoadAndDelete(reservation.RequestID)
		if !found {
			logger.Fatal().Msgf(
				"did not find reservation for request %s, crashing to reset!",
				reservation.RequestID,
			)
		}
		availableSlots.Add(slots.(int64))
	})
	if err != nil {
		logger.Fatal().Err(err).Msgf("error subscribing to scheduling.reserve.%s", utils.WORKER_ID)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
	logger.Warn().Msg("received shutdown signal!")
}

func startCoordinator(nc *nats.Conn) {
	logger.Debug().Msg("starting coordinator api")

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
