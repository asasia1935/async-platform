package main

import (
	"github.com/asasia1935/async-platform/internal/queue"
	"github.com/asasia1935/async-platform/internal/worker"
)

func main() {
	q := queue.NewQueue()

	worker.Run(q)
}
