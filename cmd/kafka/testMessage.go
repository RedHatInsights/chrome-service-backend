package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/joho/godotenv"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	godotenv.Load()
	cfg := config.Get()

	kafkaWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  cfg.KafkaConfig.KafkaBrokers,
		Topic:    cfg.KafkaConfig.KafkaTopics[0],
		Balancer: &kafka.LeastBytes{},
	})

	defer kafkaWriter.Close()

	body := `{"broadcast": true, "payload": {"Ya know what I want?": "Orange Box 3."}}`
	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("Key-%v", time.Now().Unix())),
		Value: []byte(body),
	}

	err := kafkaWriter.WriteMessages(context.TODO(), msg)
	if err != nil {
		log.Fatalf("Could not write message due to error : %v", err)
	}
}
