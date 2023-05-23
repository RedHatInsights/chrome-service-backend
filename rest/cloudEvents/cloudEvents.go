package cloudEvents

import (
	"fmt"
	"time"

	"github.com/RedHatInsights/chrome-service-backend/rest/connectionhub"
)

// Cloud events spec: https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md

type DataContentType string
type SpecVersion string

const (
	ApplicationJson DataContentType = "application/json"
	V102            SpecVersion     = "1.0.2"
)

func (dct DataContentType) IsValid() error {
	switch dct {
	case ApplicationJson:
		{
			return nil
		}
	}
	return fmt.Errorf("invalid cloud events content type, expected one of %v, got %v", []string{string(ApplicationJson)}, dct)
}

func (sv SpecVersion) IsValid() error {
	switch sv {
	case V102:
		{
			return nil
		}
	}

	return fmt.Errorf("invalid cloud events spec version, expect %v, got %v", V102, sv)
}

// TODO: Specify accepted data payload
// data type is generic, we accept any valid JSON for now
type Envelope[D any] struct {
	SpecVersion     SpecVersion     `json:"specversion"`
	Type            string          `json:"type"`
	Source          string          `json:"source"`
	Id              string          `json:"id"`
	Time            time.Time       `json:"time"`
	DataContentType DataContentType `json:"datacontenttype"`
	Data            D               `json:"data"`
}

func WrapPayload[P any](payload P, source string, id string, messageType string) Envelope[P] {
	event := Envelope[P]{
		SpecVersion:     V102,
		Type:            messageType,
		Source:          source,
		Id:              id,
		Time:            time.Now(),
		DataContentType: ApplicationJson,
		Data:            payload,
	}
	return event
}

type KafkaEnvelope struct {
	Envelope[connectionhub.WsMessage]
}
