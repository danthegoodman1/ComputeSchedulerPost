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
	}
)
