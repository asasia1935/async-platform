package queue

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/asasia1935/async-platform/internal/message"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Queue struct {
	client *redis.Client
	name   string
}

func NewQueue() *Queue {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &Queue{
		client: rdb,
		name:   "default",
	}
}

func NewDLQ() *Queue {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	return &Queue{
		client: rdb,
		name:   "default:dlq",
	}
}

func (q *Queue) Enqueue(msg message.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}

	err = q.client.LPush(ctx, q.name, data).Err()
	if err != nil {
		log.Fatal(err)
	}
}

func (q *Queue) Dequeue() (message.Message, error) {
	result, err := q.client.BRPop(ctx, 0, q.name).Result()
	if err != nil {
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
