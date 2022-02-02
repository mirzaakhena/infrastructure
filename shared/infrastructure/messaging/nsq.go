package messaging

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"infrastructure/shared/model/payload"
	"os"
	"os/signal"
	"syscall"
)

type publisherImpl struct {
	producer *nsq.Producer
}

// NewPublisher is
func NewPublisher(url string) *publisherImpl {

	nsqConfig := nsq.NewConfig()
	producer, err := nsq.NewProducer(url, nsqConfig)
	if err != nil {
		panic(err.Error())
	}

	return &publisherImpl{producer: producer}
}

// Publish is
func (m *publisherImpl) Publish(topic string, data payload.Payload) error {
	dataInBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return m.producer.Publish(topic, dataInBytes)
}

type subscriberImpl struct {
	channel     string
	subscribers map[string]*nsq.Consumer
}

func NewSubscriber(channel string) Subscriber {
	return &subscriberImpl{
		channel:     channel,
		subscribers: map[string]*nsq.Consumer{},
	}
}

func (r *subscriberImpl) Handle(topic string, onReceived HandleFunc) {

	nsqConfig := nsq.NewConfig()

	con, err := nsq.NewConsumer(topic, r.channel, nsqConfig)
	if err != nil {
		panic(err.Error())
	}

	con.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		var data payload.Payload
		err := json.Unmarshal(m.Body, &data)
		onReceived(data, err)
		return nil
	}))

	r.subscribers[topic] = con
}

func (r *subscriberImpl) Run(url string) {

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	for _, con := range r.subscribers {
		if err := con.ConnectToNSQD(url); err != nil {
			panic(err.Error())
		}
	}

	<-termChan

	for _, con := range r.subscribers {
		con.Stop()
	}
	for _, con := range r.subscribers {
		<-con.StopChan
	}
}
