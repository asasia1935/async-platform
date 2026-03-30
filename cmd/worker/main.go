package main

import (
	"github.com/asasia1935/async-platform/internal/queue"
	"github.com/asasia1935/async-platform/internal/worker"
)

// workerCount는 동시에 실행할 워커 고루틴의 수
const workerCount = 3

func main() {
	q := queue.NewQueue()

	// 여러 워커 고루틴을 실행
	for i := 0; i < workerCount; i++ {
		go worker.Run(i, q)
	}

	// 메인 고루틴이 종료되지 않도록 무한 대기
	select {}
}
