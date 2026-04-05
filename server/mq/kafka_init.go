package mq

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-payment-mservice/server/log"
)

type KafkaMsgProcessor func(msg []byte) error

func Init() {
	waitForKafka(config.Config.KafkaConfig.Brokers, 30*time.Second)
	startKafkaConsumer("user-activated", userActivationProcess)
	log.Logger.Infof("Kafka consumer for topic 'user-activated' started")
}

// waitForKafka blocks until at least one broker accepts a TCP connection
// or the timeout is reached. This prevents the consumer from silently
// failing when the broker is still initialising.
func waitForKafka(brokers []string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		for _, broker := range brokers {
			conn, err := net.DialTimeout("tcp", broker, 2*time.Second)
			if err == nil {
				conn.Close()
				// Verify we can reach the Kafka protocol (fetch metadata).
				c, err := kafka.Dial("tcp", broker)
				if err == nil {
					_, err = c.Brokers()
					c.Close()
					if err == nil {
						log.Logger.Infof("Kafka broker %s is ready", broker)
						return
					}
				}
			}
		}
		log.Logger.Warnf("Kafka brokers not ready, retrying in 2s...")
		time.Sleep(2 * time.Second)
	}
	panic(fmt.Sprintf("Kafka brokers %v not reachable after %s", brokers, timeout))
}

func startKafkaConsumer(topic string, processor KafkaMsgProcessor) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Config.KafkaConfig.Brokers,
		Topic:          topic,
		GroupID:        config.Config.KafkaConfig.GroupID,
		MaxBytes:       config.Config.KafkaConfig.MaxBytes,                      // 10MB
		CommitInterval: time.Duration(config.Config.KafkaConfig.CommitInterval), // disable auto-commit
	})
	go func() {
		for {
			ctx := context.Background()
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Logger.Errorf("Error reading message: %v", err)
				time.Sleep(time.Second)
				continue
			}
			log.Logger.Infof("Message received: Topic=%s, Key=%s, Value=%s", m.Topic, m.Key, string(m.Value))
			err = processor(m.Value)
			if err == nil {
				cmitErr := reader.CommitMessages(ctx, m)
				if cmitErr != nil {
					log.Logger.Errorf("Failed to commit message at offset %d: %v", m.Offset, cmitErr)
				}
				log.Logger.Infof("Topic: %s, Key: %s, Message at offset %d processed and committed", m.Topic, m.Key, m.Offset)
			}
		}
	}()
}
