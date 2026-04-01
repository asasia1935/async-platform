package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/asasia1935/async-platform/internal/message"
	"github.com/asasia1935/async-platform/internal/queue"
)

const maxRetry = 3

func Run(ctx context.Context, workerID int, q *queue.Queue, dlq *queue.Queue) {

	for {
		// 컨텍스트로 종료 신호가 오면 루프 탈출 -> 워커 고루틴 정상 종료
		select {
		case <-ctx.Done():
			log.Printf("worker %d shutting down", workerID)
			return
		default:
		}

		// BRPop은 블로킹 방식으로 큐에서 메시지를 꺼내옵니다.
		// 큐에 메시지가 없으면 새 메시지가 들어올 때까지 대기합니다.
		popped, err := q.Dequeue(ctx, 1*time.Second) // 1초 타임아웃 -> 메시지가 없으면 1초마다 타임아웃 에러 발생 -> 워커가 종료 신호를 체크할 수 있도록 함
		if err != nil {
			// timeout이면 정상적인 깨어남이므로 그냥 다시 루프
			if errors.Is(err, queue.ErrDequeueTimeout) {
				continue
			}

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

			// 최대 재시도 횟수를 초과할 경우
			if popped.Retry > maxRetry {
				log.Printf("worker %d max retry reached for message: %s", workerID, popped.Payload)
				// 최대 재시도 횟수를 초과한 메시지를 DLQ에 넣습니다.
				dlq.Enqueue(popped)
				continue
			}

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
