package saramax

import (
	"context"
	"encoding/json"
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
		msgs := make([]*sarama.ConsumerMessage, 0, b.BatchSize)
		data := make([]T, 0, b.BatchSize)
		done := false

		for i := 0; i < b.BatchSize && !done; i++ {
			select {
			case <-ctx.Done():
				done = true
			case msg, ok := <-msgCh:
				if !ok {
					cancel()
					return nil // 消息通道关闭，正常退出
				}
				var t T
				err := json.Unmarshal(msg.Value, &t)
				if err != nil {
					b.l.Error("反序列化失败", logger.Error(err))
					continue
				}
				msgs = append(msgs, msg)
				data = append(data, t)
			}
		}
		cancel()
		if len(data) == 0 {
			time.Sleep(1 * time.Millisecond)
			continue
		}
		go func() {
			err := b.fn(msgs, data)
			if err != nil {
				b.l.Error("batch processing failed", logger.Error(err))
			}
			for _, msg := range msgs {
				session.MarkMessage(msg, "")
			}
		}()
	}
}
