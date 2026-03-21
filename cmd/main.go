package main

import (
	"log"

	"github.com/asasia1935/async-platform/internal/message"
	"github.com/asasia1935/async-platform/internal/queue"
)

func main() {
	q := queue.NewQueue()

	msg := message.Message{
		Type:    "test",
		Payload: "hello async",
	}

	q.Enqueue(msg)

	log.Printf("enqueue: type=%s payload=%s\n", msg.Type, msg.Payload)

	popped, err := q.Dequeue()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dequeue: type=%s payload=%s\n", popped.Type, popped.Payload)
}
