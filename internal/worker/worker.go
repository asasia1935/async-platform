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
			log.Printf("level=INFO worker=%d action=worker_stopping reason=context_canceled", workerID)
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
			log.Printf("level=ERROR worker=%d action=dequeue_error queue=%s err=%v", workerID, q.Name(), err)
			// 에러가 발생하면 잠시 대기 후 재시도 (로그가 너무 많이 찍히는 것을 방지하기 위해)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		log.Printf("level=INFO worker=%d action=dequeue queue=%s type=%s payload=%q retry=%d", workerID, q.Name(), popped.Type, popped.Payload, popped.Retry)

		// 메시지 타입에 따라 적절한 핸들러로 분기 처리
		if err := dispatch(workerID, popped); err != nil {
			log.Printf("level=ERROR worker=%d action=dispatch_error queue=%s type=%s payload=%q err=%v", workerID, q.Name(), popped.Type, popped.Payload, err)

			// 재시도 횟수 증가
			popped.Retry++

			// 최대 재시도 횟수를 초과할 경우
			if popped.Retry > maxRetry {
				log.Printf("level=ERROR worker=%d action=max_retry_reached queue=%s type=%s payload=%q", workerID, q.Name(), popped.Type, popped.Payload)

				// 최대 재시도 횟수를 초과한 메시지를 DLQ에 넣습니다.
				dlq.Enqueue(ctx, popped)
				log.Printf("level=WARN worker=%d action=move_to_dlq queue=%s type=%s payload=%q retry=%d reason=max_retry_exceeded", workerID, dlq.Name(), popped.Type, popped.Payload, popped.Retry)

				continue
			}

			// 에러가 발생하면 해당 메시지를 큐에 다시 넣어서 재시도 합니다.
			q.Enqueue(ctx, popped)

			log.Printf("level=WARN worker=%d action=retry_enqueue queue=%s type=%s payload=%q retry=%d",
				workerID, q.Name(), popped.Type, popped.Payload, popped.Retry)
		}
	}
}

func dispatch(workerID int, msg message.Message) error {
	switch msg.Type {
	case "test":
		return handleTest(workerID, msg)
	default:
		return ErrUnknownMessageType
	}
}

func handleTest(workerID int, msg message.Message) error {
	log.Printf("level=INFO worker=%d action=handle_test_start payload=%q", workerID, msg.Payload)

	// 실제 worker pool을 실험하기 위해 handleTest에서 지연 추가
	time.Sleep(2 * time.Second)

	log.Printf("level=INFO worker=%d action=handle_test_done payload=%q", workerID, msg.Payload)

	// 테스트를 위해 특정 페이로드에서 에러를 발생시키도록 함 -> 재시도 로직과 DLQ 이동 로직이 정상적으로 동작하는지 확인하기 위함 (에러 메시지 정의 X)
	if msg.Payload == "hello async 5" {
		return errors.New("simulated error for payload: " + msg.Payload)
	}

	return nil
}
