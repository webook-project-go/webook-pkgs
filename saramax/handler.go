package saramax

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/webook-project-go/webook-pkgs/logger"
)

type Handler[T any] struct {
	l  logger.Logger
	fn func(msg *sarama.ConsumerMessage, value T) error
}

func NewHandler[T any](l logger.Logger, fn func(msg *sarama.ConsumerMessage, value T) error) *Handler[T] {
	return &Handler[T]{
		l:  l,
		fn: fn,
	}
}
func (h *Handler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		var t T
		err := json.Unmarshal(msg.Value, &t)
		if err != nil {
			session.MarkMessage(msg, "")
			continue
		}
		for i := 0; i < 3; i++ {
			err = h.fn(msg, t)
			if err != nil {
				continue
			}
			break
		}
		if err != nil {
			h.l.Error("处理消息失败， 重试上限", logger.Error(err),
				logger.String("topic", msg.Topic),
				logger.Int64("offset", msg.Offset),
				logger.Int32("partition", msg.Partition),
			)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
