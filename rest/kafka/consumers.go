package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/cloudevents"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
	"log"
	"os"
)

type kafkaConsumer struct {
	Topics  []string
	Readers map[string]*kafka.Reader
}

var Consumer = kafkaConsumer{}

const TenMb = 10e7

func createReader(topic string) *kafka.Reader {
	cfg := config.Get()
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorln("Couldn't get hostname, using UUID")
		hostname = uuid.NewString()
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.KafkaConfig.KafkaBrokers,
		// The consumer group will match the pod name via hostname
		// ex platform.chrome.chrome-service-api.<deployHash>.<podHash>
		GroupID:     fmt.Sprintf("platform.chrome.%s", hostname),
		StartOffset: kafka.LastOffset,
		Topic:       topic,
		Logger:      kafka.LoggerFunc(logrus.Debugf),
		ErrorLogger: kafka.LoggerFunc(logrus.Errorf),
		MaxBytes:    TenMb,
	})
	logrus.Infoln("Creating new kafka reader for topic:", topic)
	return r
}

func InitializeConsumers() {
	cfg := config.Get()
	topics := cfg.KafkaConfig.KafkaTopics
	readers := make(map[string]*kafka.Reader)
	for _, topic := range topics {
		readers[topic] = createReader(topic)
	}

	Consumer.Readers = readers

	for _, r := range readers {
		go startKafkaReader(r)
	}
}

func startKafkaReader(r *kafka.Reader) {
	defer r.Close()
	for {
		logrus.Infoln("Reading from ", r.Config().Topic)
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			logrus.Errorln("Error reading message: ", err)
			break
		}
		logrus.Infoln(fmt.Sprintf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value)))

		var p cloudevents.KafkaEnvelope
		err = json.Unmarshal(m.Value, &p)
		if err != nil {
			logrus.Errorln(fmt.Sprintf("Unable to unmarshal message %s\n", string(m.Value)))
		} else if p.Data.Payload == nil {
			logrus.Errorln(fmt.Sprintf("No message will be emitted do to missing payload %s! Message might not follow cloud events spec.\n", string(m.Value)))
		} else {
			event := cloudevents.WrapPayload(p.Data.Payload, p.Source, p.Id, p.Type)
			event.Time = p.Time
			data, err := json.Marshal(event)
			if err != nil {
				log.Println("Unable to marshal payload data", p, err)
			} else {
				validateErr := cloudevents.ValidatePayload(p)
				if validateErr == nil {
					newMessage := connectionhub.Message{
						Destinations: connectionhub.MessageDestinations{
							Users:         p.Data.Users,
							Roles:         p.Data.Roles,
							Organizations: p.Data.Organizations,
						},
						Broadcast: p.Data.Broadcast,
						Data:      data,
					}
					if p.Data.Broadcast {
						logrus.Infoln("Emitting new broadcast message from kafka reader: ", string(newMessage.Data))
						connectionhub.ConnectionHub.Broadcast <- newMessage
					} else {
						logrus.Infoln("Emitting new message from kafka reader: ", string(newMessage.Data))
						connectionhub.ConnectionHub.Emit <- newMessage
					}
				} else {
					logrus.Errorln(validateErr)
				}
			}
		}
	}

	if err := r.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
}
