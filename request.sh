curl -d '{
  "Task": "increment",
  "Payload": {
    "Num": 1
  },
  "Requirements": {
    "Slots": 5,
    "Region": "us-east"
  }
}' -H 'Content-Type: application/json' http://localhost:8080/schedule