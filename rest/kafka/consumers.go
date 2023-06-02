package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/cloudevents"
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

		var p cloudevents.KafkaEnvelope
		err = json.Unmarshal(m.Value, &p)
		if err != nil {
			log.Printf("Unable Unmarshal message %s\n", string(m.Value))
		} else if p.Data.Payload == nil {
			log.Printf("No message will be emitted doe to missing payload %s! Message might not follow cloud events spec.\n", string(m.Value))
		} else {
			event := cloudevents.WrapPayload(p.Data.Payload, p.Source, p.Id, p.Type)
			event.Time = p.Time
			data, err := json.Marshal(event)
			if err != nil {
				log.Println("Unable marshal payload data", p, err)
			} else {
				err := p.DataContentType.IsValid()
				if err != nil {
					log.Println("Kafka message payload needs to be JSON formatted", err)
				} else {
					err := p.SpecVersion.IsValid()
					if err != nil {
						log.Println(err)
					} else {
						err := p.Source.IsValid()
						if err != nil {
							log.Println(err)
						} else {
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
						}
					}
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
		Logger:      kafka.LoggerFunc(logrus.Debugf),
		ErrorLogger: kafka.LoggerFunc(logrus.Errorf),
		MinBytes:    1,    // 1B
		MaxBytes:    10e7, // 10MB
	})
	logrus.Infoln("Creating new kafka reader for topic:", topic)
	return r
}

func InitializeConsumers() {
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
