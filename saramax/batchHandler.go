package saramax

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/webook-project-go/webook-pkgs/logger"
	"time"
)

type BatchHandler[T any] struct {
	BatchSize int
	TimeOut   time.Duration
	fn        func([]*sarama.ConsumerMessage, []T) error
	l         logger.Logger
}

func NewBatchHandler[T any](batchSize int, timeout time.Duration, fn func([]*sarama.ConsumerMessage, []T) error,
	l logger.Logger) *BatchHandler[T] {
	return &BatchHandler[T]{
		BatchSize: batchSize,
		TimeOut:   timeout,
		fn:        fn,
		l:         l,
	}
}
func (b *BatchHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (b *BatchHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgCh := claim.Messages()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), b.TimeOut)
		done := false
		msgs := make([]*sarama.ConsumerMessage, 0, b.BatchSize)
		data := make([]T, 0, b.BatchSize)
		var last *sarama.ConsumerMessage
		for i := 0; i < b.BatchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					cancel()
					return errors.New("message chan closed")
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化失败")
					continue
				}
				msgs = append(msgs, msg)
				data = append(data, t)
				last = msg
			}
		}
		cancel()
		_ = b.fn(msgs, data)
		if last != nil {
			session.MarkMessage(last, "")
		}
	}
}
