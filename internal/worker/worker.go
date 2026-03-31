package worker

import (
	"errors"
	"log"
	"time"

	"github.com/asasia1935/async-platform/internal/message"
	"github.com/asasia1935/async-platform/internal/queue"
)

func Run(workerID int, q *queue.Queue) {

	for {
		// BRPop은 블로킹 방식으로 큐에서 메시지를 꺼내옵니다.
		// 큐에 메시지가 없으면 새 메시지가 들어올 때까지 대기합니다.
		popped, err := q.Dequeue()
		if err != nil {
			// Worker가 계속 실행되도록 에러를 로그로 남기고 루프를 계속합니다.
			log.Printf("worker %d dequeue error: %v", workerID, err)
			continue
		}

		log.Printf("worker %d dequeue: type=%s payload=%s\n", workerID, popped.Type, popped.Payload)

		// 메시지 타입에 따라 적절한 핸들러로 분기 처리
		if err := dispatch(workerID, popped); err != nil {
			log.Printf("worker %d dispatch error: %v", workerID, err)

			// 재시도 횟수 증가
			popped.Retry++

			// 에러가 발생하면 해당 메시지를 큐에 다시 넣어서 재시도 합니다.
			q.Enqueue(popped)

			log.Printf("worker %d retrying message (retry=%d): %s",
				workerID, popped.Retry, popped.Payload)
		}
	}
}

func dispatch(workerID int, msg message.Message) error {
	switch msg.Type {
	case "test":
		return handleTest(workerID, msg)
	default:
		return errors.New("unknown message type")
	}
}

func handleTest(workerID int, msg message.Message) error {
	log.Printf("worker %d handleTest: payload=%s\n", workerID, msg.Payload)

	// 실제 worker pool을 실험하기 위해 handleTest에서 지연 추가
	time.Sleep(2 * time.Second)

	log.Printf("worker %d handleTest Done: payload=%s\n", workerID, msg.Payload)

	if msg.Payload == "hello async 5" {
		return errors.New("simulated error for payload: " + msg.Payload)
	}

	return nil
}
