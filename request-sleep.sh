curl -d '{
  "Task": "increment",
  "Payload": {
    "Num": 1,
    "SleepSec": 10
  },
  "Requirements": {
    "Region": "us-east",
    "Slots": 8
  }
}' -H 'Content-Type: application/json' http://localhost:8080/schedule