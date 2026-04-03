package queue

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/asasia1935/async-platform/internal/message"
	"github.com/redis/go-redis/v9"
)

var ErrDequeueTimeout = errors.New("dequeue timeout")

// 임의로 넣을 context 생성 -> 실제로는 main에서 context를 생성해서 전달하는 방식으로 변경할 예정
var ctx = context.Background()

type Queue struct {
	client *redis.Client
	name   string
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func NewQueue(client *redis.Client, name string) *Queue {
	return &Queue{
		client: client,
		name:   name,
	}
}

func (q *Queue) Enqueue(msg message.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	err = q.client.LPush(ctx, q.name, data).Err()
	if err != nil {
		return err
	}

	return nil
}

func (q *Queue) Dequeue(ctx context.Context, timeout time.Duration) (message.Message, error) {
	result, err := q.client.BRPop(ctx, timeout, q.name).Result()
	if err != nil {
		// context cancel이면 그대로 반환
		if errors.Is(err, context.Canceled) {
			return message.Message{}, err
		}
		// timeout이면 별도 sentinel error 반환
		if err.Error() == "redis: nil" {
			return message.Message{}, ErrDequeueTimeout
		}
		return message.Message{}, err
	}

	// BRPop 결과:
	// [0] = queue name
	// [1] = popped value
	if len(result) != 2 {
		return message.Message{}, errors.New("invalid BRPOP result")
	}

	var msg message.Message
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return message.Message{}, err
	}

	return msg, nil
}
