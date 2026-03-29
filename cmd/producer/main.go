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

	/* BRPop은 블로킹 방식으로 큐에서 메시지를 꺼내옵니다.
	   큐에 메시지가 없으면 새 메시지가 들어올 때까지 대기합니다. */
	/* 추후 worker에서 사용할 예정이므로, 일단 주석 처리합니다.
	popped, err := q.Dequeue()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("dequeue: type=%s payload=%s\n", popped.Type, popped.Payload)
	*/
}
