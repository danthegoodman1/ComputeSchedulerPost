package scheduling

type (
	Requirements struct {
		// Fake resource
		FooPoints int
	}

	ScheduleRequest struct {
		RequestID    string
		Task         string
		Requirements Requirements
	}

	ScheduleResponse struct {
		WorkerID string
		Payload  map[string]any
	}

	ReleaseResourcesMessage struct {
		RequestID string
		// The worker that may ignore this. Set to empty string to force all workers to release
		ExemptWorker string
	}

	ReserveRequest struct {
		Task    string
		Payload map[string]any
	}

	ReserveResponse struct {
		Error   *string
		Payload map[string]any
	}
)
