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

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error is not retryable",
			err:  nil,
			want: false,
		},
		{
			name: "unknown message type is not retryable",
			err:  ErrUnknownMessageType,
			want: false,
		},
		{
			name: "generic error is retryable",
			err:  errors.New("temporary handler failure"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRetryableError(tt.err); got != tt.want {
				t.Fatalf("isRetryableError() = %v, want %v", got, tt.want)
			}
		})
	}
}
