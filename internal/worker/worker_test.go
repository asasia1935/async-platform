package worker

import (
	"errors"
	"testing"

	"github.com/asasia1935/async-platform/internal/message"
)

func TestDispatchTestSuccessReturnsNil(t *testing.T) {
	msg := message.Message{
		Type:    "test.success",
		Payload: "success payload",
	}

	if err := dispatch(1, msg); err != nil {
		t.Fatalf("dispatch() error = %v, want nil", err)
	}
}

func TestDispatchTestFailReturnsError(t *testing.T) {
	msg := message.Message{
		Type:    "test.fail",
		Payload: "fail payload",
	}

	if err := dispatch(1, msg); err == nil {
		t.Fatal("dispatch() error = nil, want non-nil")
	}
}

func TestDispatchUnknownTypeReturnsErrUnknownMessageType(t *testing.T) {
	msg := message.Message{
		Type:    "unknown.type",
		Payload: "unknown payload",
	}

	err := dispatch(1, msg)
	if !errors.Is(err, ErrUnknownMessageType) {
		t.Fatalf("dispatch() error = %v, want ErrUnknownMessageType", err)
	}
}
