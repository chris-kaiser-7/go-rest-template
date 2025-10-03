package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const topicName = "test"

func (app *application) kafSend(msg []byte) error {
	topic := topicName
	err := app.kafkaProducer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          msg},
		nil, // delivery channel
	)
	return err
}
