package messaging

import (
	"infrastructure/shared/model/payload"
)

type Publisher interface {
	Publish(topic string, payload payload.Payload) error
}

type HandleFunc func(payload payload.Payload, err error)

type Subscriber interface {
	Handle(topic string, onReceived HandleFunc)
	Run(url string)
}
