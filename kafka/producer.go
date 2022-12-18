package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/rs/zerolog/log"
)

type Producer struct {
	p        *kafka.Producer
	servers  string
	topic    string
	stopChan chan struct{}
}

func NewProducer(servers, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": servers,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Err(err).Msgf("delivery failed: %v", ev.TopicPartition)
				} else {
					log.Debug().Msgf("produced message to %v", ev.TopicPartition)
				}
			case kafka.Error:
				log.Err(e.(kafka.Error)).Msg("kafka error")
			}
		}
	}()
	return &Producer{
		p:        p,
		topic:    topic,
		stopChan: make(chan struct{}),
	}, nil
}

func (s *Producer) Produce(message string) {
	var err error
	retry := true
	for {
		select {
		case <-s.stopChan:
			if retry {
				log.Err(err).Msgf("Failed to produce message to the topic: %s:%s", s.topic, message)
			}
			return
		default:
			err = s.p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
				Value:          []byte(message),
			}, nil)
			if err == nil {
				return
			}
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrQueueFull {
					// Producer queue is full, wait 1s for messages
					// to be delivered then try again.
					time.Sleep(time.Second)
					retry = true
					continue
				}
				log.Err(err).Msg("Failed to produce message")
				time.Sleep(2 * time.Second)
				retry = true
				continue
			}
		}
	}
}

func (s *Producer) Stop() {
	log.Info().Msg("Waiting Producer...")
	close(s.stopChan)
	s.p.Close()
	log.Info().Msg("Producer stopped")
}
