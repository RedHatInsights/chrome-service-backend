package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type kafkaConsumer struct {
	Topics  []string
	Readers map[string]*kafka.Reader
}

var KafkaConsumer = kafkaConsumer{}

func startKafkaReader(r *kafka.Reader) {
	defer r.Close()
	for {
		logrus.Infoln("Reading from ", r.Config().Topic)
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logrus.Errorln("Error reading message: ", err)
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))

		var p connectionhub.WsMessage
		err = json.Unmarshal(m.Value, &p)
		if err != nil {
			log.Printf("Unable Unmarshal message %s\n", string(m.Value))
		} else {
			data, err := json.Marshal(&p.Payload)
			if err != nil {
				log.Println("Unable marshal payload data", p, err)
			} else {
				newMessage := connectionhub.Message{
					Destinations: connectionhub.MessageDestinations{
						Users:         p.Users,
						Roles:         p.Roles,
						Organizations: p.Organizations,
					},
					Broadcast: p.Broadcast,
					Data:      data,
				}
				if p.Broadcast {
					logrus.Infoln("Emitting new broadcast message from kafka reader: ", string(newMessage.Data))
					connectionhub.ConnectionHub.Broadcast <- newMessage
				} else {
					logrus.Infoln("Emitting new message from kafka reader: ", string(newMessage.Data))
					connectionhub.ConnectionHub.Emit <- newMessage
				}
			}
		}
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}

func createReader(topic string) *kafka.Reader {
	config := config.Get()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     config.KafkaConfig.KafkaBrokers,
		Topic:       topic,
		Logger:      kafka.LoggerFunc(logrus.Infof),
		ErrorLogger: kafka.LoggerFunc(logrus.Errorf),
		MinBytes:    1,    // 1B
		MaxBytes:    10e7, // 10MB
	})
	logrus.Infoln("Creating new kafka reader for topic:", topic)
	return r
}

func InitializeConzumers() {
	config := config.Get()
	topics := config.KafkaConfig.KafkaTopics
	readers := make(map[string]*kafka.Reader)
	for _, topic := range topics {
		readers[topic] = createReader(topic)
	}

	KafkaConsumer.Readers = readers

	for _, r := range readers {
		go startKafkaReader(r)
	}
}
