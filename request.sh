curl -d '{
  "Task": "increment",
  "Payload": {
    "Num": 1,
    "SleepSec": 3
  },
  "Requirements": {
    "Slots": 5
  }
}' -H 'Content-Type: application/json' http://localhost:8080/schedule