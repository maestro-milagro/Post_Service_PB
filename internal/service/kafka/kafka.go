package kafka

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

type KafkaProducer struct {
	log      *slog.Logger
	producer *kafka.Producer
}

func New(log *slog.Logger) *KafkaProducer {
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		// User-specific properties that you must set
		"bootstrap.servers": "localhost:8082",

		// Fixed properties
		"acks": "all"})

	if err != nil {
		fmt.Printf("Failed to create producer: %s", err)
		os.Exit(1)
	}
	return &KafkaProducer{
		log:      log,
		producer: p,
	}
}

func (kf *KafkaProducer) Produce(email string, subIds []int) error {
	go func() {
		for e := range kf.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					fmt.Printf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					fmt.Printf("Produced event to topic %s: key = %-10s value = %s\n",
						*ev.TopicPartition.Topic, string(ev.Key), string(ev.Value))
				}
			}
		}
	}()
	topic := "posts"
	var strIds string
	for _, id := range subIds {
		strIds += strconv.Itoa(id) + ","
	}
	strIds = strings.Trim(strIds, ",")

	kf.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(email),
		Value:          []byte(strIds),
	}, nil)

	// Wait for all messages to be delivered
	kf.producer.Flush(15 * 1000)
	kf.producer.Close()
}
