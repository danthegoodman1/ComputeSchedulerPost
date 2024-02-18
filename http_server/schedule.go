package http_server

import (
	"context"
	"errors"
	"fmt"
	"github.com/danthegoodman1/GoAPITemplate/scheduling"
	"github.com/danthegoodman1/GoAPITemplate/utils"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

type PostScheduleRequest struct {
	Task         string // what task we are scheduling
	Requirements scheduling.Requirements
	Payload      map[string]any // task payload
}

func (s *HTTPServer) PostSchedule(c *CustomContext) error {
	var body PostScheduleRequest
	if err := ValidateRequest(c, &body); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	// Request resources from workers
	logger.Debug().Msgf("Asking workers to schedule '%s'", body.Task)
	msg, err := s.NatsClient.Request(fmt.Sprintf("scheduling.request.%s", body.Task), utils.JSONMustMarshal(scheduling.ScheduleRequest{
		RequestID:    c.RequestID, // use the unique request ID
		Task:         body.Task,
		Requirements: body.Requirements,
	}), time.Second*5)

	if errors.Is(err, context.DeadlineExceeded) {
		s.mustEmitRelease(c.RequestID, "") // tell all workers to release resources
		return echo.NewHTTPError(http.StatusGatewayTimeout, "no workers responded in time")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var workerRes scheduling.ScheduleResponse
	utils.JSONMustUnmarshal(msg.Data, &workerRes)

	logger.Debug().Msgf("Worker %s responded to schedule request", workerRes.WorkerID)
	// tell other workers to release resources
	s.mustEmitRelease(c.RequestID, workerRes.WorkerID)

	// Tell the worker that they are reserved, and to do the task
	msg, err = s.NatsClient.Request(fmt.Sprintf("scheduling.reserve_task.%s", workerRes.WorkerID), utils.JSONMustMarshal(scheduling.ReserveRequest{
		Task:    body.Task,
		Payload: body.Payload,
	}), time.Second*5)

	var reserveRes scheduling.ReserveResponse
	utils.JSONMustUnmarshal(msg.Data, &reserveRes)

	if reserveRes.Error != nil {
		c.String(http.StatusInternalServerError, *reserveRes.Error)
	}

	return c.JSON(http.StatusOK, reserveRes.Payload)
}

func (s *HTTPServer) mustEmitRelease(requestID, exemptWorker string) {
	// emit special cancel topic
	err := s.NatsClient.Publish("scheduling.release", utils.JSONMustMarshal(scheduling.ReleaseResourcesMessage{
		RequestID:    requestID,
		ExemptWorker: exemptWorker,
	}))
	if err != nil {
		panic(err)
	}
}
