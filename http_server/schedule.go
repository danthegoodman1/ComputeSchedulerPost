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
	Payload      *map[string]any // task payload
}

func (s *HTTPServer) PostSchedule(c *CustomContext) error {
	var body PostScheduleRequest
	if err := ValidateRequest(c, &body); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	logger.Debug().Msgf("Asking workers to schedule '%s'", body.Task)
	msg, err := s.NatsClient.Request(fmt.Sprintf("workers.%s", body.Task), utils.JSONMustMarshal(scheduling.ScheduleRequest{
		RequestID:    c.RequestID, // use the unique request ID
		Task:         body.Task,
		Requirements: body.Requirements,
	}), time.Second*5)
	if errors.Is(err, context.DeadlineExceeded) {
		s.mustEmitCancel(c.RequestID) // tell the nodes to release resources
		return echo.NewHTTPError(http.StatusGatewayTimeout, "no workers responded in time")
	} else if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	var workerRes scheduling.ScheduleResponse
	utils.JSONMustUnmarshal(msg.Data, &workerRes)

	logger.Debug().Msgf("Worker %s responded to schedule request")

	s.mustEmitCancel(c.RequestID) // tell all nodes to release resources, the responding node will ignore this request

	return c.JSON(http.StatusOK, workerRes.Payload)
}

func (s *HTTPServer) mustEmitCancel(requestID string) {
	// emit special cancel topic
	err := s.NatsClient.Publish("workers._cancel", utils.JSONMustMarshal(scheduling.ReleaseResourcesMessage{RequestID: requestID}))
	if err != nil {
		panic(err)
	}
}
