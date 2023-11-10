package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/google/uuid"
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

	id := uuid.New()
	body := fmt.Sprintf(`{
		"specversion": "1.0.2",
		"type": "notifications.drawer",
		"source": "https://whatever.service.com",
		"id": "test-message",
		"time": "2023-05-23T11:54:03.879689005+02:00",
		"datacontenttype": "application/json",
		"data":{
			"broadcast": false,
			"organizations": ["11789772"],
			"payload": {
				"id": "%s",
				"description": "string",
				"title": "string",
				"created": "2023-05-23T11:54:03.879689005+02:00",
				"read": false,
				"source": "string"
			}
		}
	}`, id.String())

	msg := kafka.Message{
		Key:   []byte(fmt.Sprintf("Key-%v", time.Now().Unix())),
		Value: []byte(body),
	}

	fmt.Println("Targeting Topic: ", kafkaWriter.Topic)
	fmt.Println("Sending message", body)

	err := kafkaWriter.WriteMessages(context.TODO(), msg)
	if err != nil {
		log.Fatalf("Could not write message due to error : %v", err)
	}
}
