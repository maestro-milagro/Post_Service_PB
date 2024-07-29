package kafka

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/maestro-milagro/Post_Service_PB/internal/models"
	"log"
	"log/slog"
)

type KafkaProducer struct {
	log      *slog.Logger
	producer sarama.AsyncProducer
}

func New(log *slog.Logger, brokers []string) *KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	return &KafkaProducer{
		log:      log,
		producer: producer,
	}
}

func prepareMessage(topic string, message []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(message),
	}
	return msg
}

func (kf *KafkaProducer) Produce(post models.Post, topic string) {
	go func() {
		for err := range kf.producer.Errors() {
			kf.log.Error("async producer error:", err.Err)
		}
	}()

	go func() {
		for succ := range kf.producer.Successes() {
			kf.log.Info("async producer success:", succ)
		}
	}()

	postJson, err := json.Marshal(post)
	if err != nil {
		log.Fatal(err)
	}

	msg := prepareMessage(topic, postJson)

	kf.producer.Input() <- msg

	kf.log.Info("post sent %v: ", post.PostID)
}
