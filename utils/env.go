package utils

import "os"

var (
	NATS_URL  = os.Getenv("NATS_URL")
	WORKER_ID = os.Getenv("WORKER_ID")
)
