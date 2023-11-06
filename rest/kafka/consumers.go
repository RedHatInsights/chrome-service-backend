package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/config"
	"github.com/RedHatInsights/chrome-service-backend/rest/cloudevents"
	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
	"github.com/google/uuid"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/sirupsen/logrus"
)

type kafkaConsumer struct {
	Topics  []string
	Readers map[string]*kafka.Reader
}

var Consumer = kafkaConsumer{}

const (
	TenMb       = 10e7
	saslPlain   = "plain"
	scramSha256 = "scram-sha-256"
	scramSha512 = "scram-sha-512"
)

var SaslMechanism sasl.Mechanism

func CreateSaslMechanism(saslConfig *clowder.KafkaSASLConfig) (sasl.Mechanism, error) {
	if SaslMechanism != nil {
		return SaslMechanism, nil
	}

	if saslConfig == nil {
		return nil, errors.New("could not create a Sasl mechanism for Kafka: the Sasl configuration settings are empty")
	}

	if saslConfig.SaslMechanism == nil || *saslConfig.SaslMechanism == "" {
		return nil, errors.New("could not create a Sasl mechanism for Kafka: the Sasl mechanism is empty")
	}

	if saslConfig.Username == nil {
		return nil, errors.New("could not create a Sasl mechanism for Kafka: the Sasl username is nil")
	}

	if saslConfig.Password == nil {
		return nil, errors.New("could not create a Sasl mechanism for Kafka: the Sasl password is nil")
	}

	var algorithm scram.Algorithm
	switch strings.ToLower(*saslConfig.SaslMechanism) {
	case saslPlain:
		return plain.Mechanism{Username: *saslConfig.Username, Password: *saslConfig.Password}, nil
	case scramSha256:
		algorithm = scram.SHA256
	case scramSha512:
		algorithm = scram.SHA512
	default:
		return nil, fmt.Errorf(`unable to configure Sasl mechanism "%s" for Kafka`, *saslConfig.SaslMechanism)
	}

	mechanism, err := scram.Mechanism(algorithm, *saslConfig.Username, *saslConfig.Password)
	if err != nil {
		return nil, fmt.Errorf(`unable to generate "%s" mechanism with the "%v" algorithm: %s`, *saslConfig.SaslMechanism, algorithm, err)
	}

	SaslMechanism = mechanism
	return SaslMechanism, nil
}

func createReader(topic string) *kafka.Reader {
	cfg := config.Get()
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Errorln("Couldn't get hostname, using UUID")
		hostname = uuid.NewString()
	}

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	mechanism, err := CreateSaslMechanism(cfg.KafkaConfig.BrokerConfig.Sasl)
	if err == nil {
		dialer.SASLMechanism = mechanism
	} else {
		logrus.Errorln("Couldn't create SASL mechanism for Kafka: ", err)
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
		Dialer:      dialer,
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
			logrus.Errorln(fmt.Sprintf("No message will be emitted due to missing payload %s! Message might not follow cloud events spec.\n", string(m.Value)))
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
