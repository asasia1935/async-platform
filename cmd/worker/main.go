package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/asasia1935/async-platform/internal/queue"
	"github.com/asasia1935/async-platform/internal/worker"
)

// workerCount는 동시에 실행할 워커 고루틴의 수
const workerCount = 3

func main() {
	// 종료 신호를 전달할 context 생성
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Redis 클라이언트 생성 -> 큐와 DLQ(Dead Letter Queue)에서 재사용
	rdb := queue.NewRedisClient("localhost:6379")

	// 큐와 DLQ(Dead Letter Queue) 생성
	q := queue.NewQueue(rdb, "default")
	dlq := queue.NewQueue(rdb, "default:dlq")

	// OS 종료 신호 수신 채널
	sigCh := make(chan os.Signal, 1)                    // OS 신호를 받을 채널 생성 -> 버퍼 크기를 1로 설정하여 Block 없이 신호를 받을 수 있도록 함
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM) // 특정 OS 신호(인터럽트, 프로세스 종료)를 sigCh 채널로 전달하도록 설정

	// 워커 고루틴을 실행하기 위한 WaitGroup 생성 : 워커 고루틴이 종료될 때까지 기다리는 용도로 사용
	var wg sync.WaitGroup

	// 종료 신호를 받으면 cancel 호출
	go func() {
		sig := <-sigCh
		log.Printf("level=INFO action=shutdown_signal signal=%q", sig)
		cancel()
	}()

	// 여러 워커 고루틴을 실행
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			// 워커 고루틴이 종료(cancel 호출)될 때 WaitGroup에서 Done 호출 -> main goroutine이 모든 워커가 종료될 때까지 대기할 수 있도록 함
			defer wg.Done()

			// 워커 고루틴 실행 -> Run 함수에서 ctx.Done()을 체크하여 종료 신호를 받으면 루프 탈출
			worker.Run(ctx, workerID, q, dlq)
		}(i)
	}

	// main goroutine도 종료 신호를 기다림
	<-ctx.Done() // ctx.Done 채널이 닫힐때까지 대기 -> cancel()이 호출되면 ctx.Done 채널이 닫히면서 main goroutine이 종료 신호를 받음
	log.Println("level=INFO action=shutdown_started")

	wg.Wait() // 모든 워커 고루틴이 종료될 때까지 대기
	log.Println("level=INFO action=all_workers_stopped")

	log.Println("level=INFO action=worker_main_shutdown_complete")
}
