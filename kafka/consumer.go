package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
)

type Consumer struct {
	c           *kafka.Consumer
	wg          sync.WaitGroup
	batchSize   int
	flushPeriod time.Duration
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewConsumer(servers, groupId string, topics []string, offset string) (*Consumer, error) {
	sessionTimeoutMs := 60000
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     servers,
		"broker.address.family": "v4",
		"group.id":              groupId,
		// We will commit the offset manually.
		"enable.auto.commit": false,
		// The session.timeout should be longer than batch processing time
		// default 45 sec
		"session.timeout.ms": sessionTimeoutMs,
		// Logical offsets: "beginning", "earliest", "end", "latest", "unset", "invalid", "stored"
		"auto.offset.reset": offset,
	})

	if err != nil {
		return nil, err
	}

	err = c.SubscribeTopics(topics, nil)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Consumer{
		c:           c,
		ctx:         ctx,
		cancel:      cancel,
		batchSize:   10,
		flushPeriod: 5 * time.Second,
	}, nil
}

// Consume does the batch message processing.
// The handleBatch triggered when the batchSize reached or the flushPeriod reached
func (c *Consumer) Consume(handleFunc func(ctx context.Context, msg []string) error) {
	c.wg.Add(1)
	println(c.flushPeriod.Seconds())
	go func(c *Consumer) {
		defer c.wg.Done()
		messageBatch := make([]*kafka.Message, 0, c.batchSize)
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-c.ctx.Done():
				log.Debug().Msg("return from the consumer")
				return
			case tick := <-timeout:
				// Process the message batch once the flushPeriod reached but a batch is not full
				log.Debug().Msgf("flush period reached %s.", tick)
				if len(messageBatch) > 0 {
					c.handleBatch(c.ctx, handleFunc, messageBatch)
					messageBatch = make([]*kafka.Message, 0, c.batchSize)
				}
				timeout = time.After(c.flushPeriod)
			default:
				msg, err := c.c.ReadMessage(c.flushPeriod)
				if err == nil {
					log.Debug().Msgf("consumed message on %s", msg.TopicPartition)
					messageBatch = append(messageBatch, msg)
					// Process the message batch once the batchSize reached
					if len(messageBatch) >= c.batchSize {
						c.handleBatch(c.ctx, handleFunc, messageBatch)
						messageBatch = make([]*kafka.Message, 0, c.batchSize)
					}
				} else if err.(kafka.Error).Code() != kafka.ErrTimedOut {
					// The client will automatically try to recover from all errors.
					// Timeout is not considered an error because it is raised by
					// ReadMessage in absence of messages.
					log.Err(err).Msg("consumer error")
				}
			}
		}
	}(c)
}

func (c *Consumer) handleBatch(ctx context.Context, handleFunc func(ctx context.Context, msg []string) error, batch []*kafka.Message) {
	// Collect the message values'
	msgs := make([]string, len(batch))
	for k := range batch {
		if batch[k] == nil {
			continue
		}
		msgs[k] = string(batch[k].Value)
	}
	// Apply the handleFunc
	err := handleFunc(ctx, msgs)
	if errors.Is(err, context.Canceled) {
		// Do not mark messages offset if consumer stopped
		log.Info().Msg("exit handler by the context cancellation")
		return
	}
	// Mark offset
	for k := range batch {
		_, err := c.c.CommitMessage(batch[k])
		if err != nil {
			log.Err(err).Msg("message committing error")
		}
	}
}

func (c *Consumer) Stop() {
	log.Info().Msg("waiting handler...")
	c.cancel()
	c.wg.Wait()
	err := c.c.Close()
	if err != nil {
		log.Err(err).Msg("consumer error")
	}
	log.Info().Msg("consumer stopped")
}
