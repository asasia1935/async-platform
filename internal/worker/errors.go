package worker

import "errors"

var (
	ErrUnknownMessageType = errors.New("unknown message type")
	ErrTaskFailed         = errors.New("task execution failed")
)
