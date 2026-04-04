package queue

import "errors"

var (
	ErrDequeueTimeout     = errors.New("dequeue timeout")
	ErrInvalidQueueResult = errors.New("invalid queue result")
)
