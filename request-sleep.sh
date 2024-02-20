curl -d '{
  "Task": "increment",
  "Payload": {
    "Num": 1,
    "SleepSec": 4
  },
  "Requirements": {
    "Slots": 8
  }
}' -H 'Content-Type: application/json' http://localhost:8080/schedule