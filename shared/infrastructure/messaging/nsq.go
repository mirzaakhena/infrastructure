package messaging

import (
	"encoding/json"
	"github.com/nsqio/go-nsq"
	"infrastructure/shared/model/payload"
	"os"
	"os/signal"
	"syscall"
)

type publisherNSQImpl struct {
	producer *nsq.Producer
}

// NewPublisherNSQ is
func NewPublisherNSQ(url string) *publisherNSQImpl {

	nsqConfig := nsq.NewConfig()
	producer, err := nsq.NewProducer(url, nsqConfig)
	if err != nil {
		panic(err.Error())
	}

	return &publisherNSQImpl{producer: producer}
}

// Publish is
func (m *publisherNSQImpl) Publish(topic string, data payload.Payload) error {
	dataInBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return m.producer.Publish(topic, dataInBytes)
}

type subscriberNSQImpl struct {
	channel     string
	subscribers map[string]*nsq.Consumer
}

func NewSubscriberNSQ(channel string) Subscriber {
	return &subscriberNSQImpl{
		channel:     channel,
		subscribers: map[string]*nsq.Consumer{},
	}
}

func (r *subscriberNSQImpl) Handle(topic string, onReceived HandleFunc) {

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

func (r *subscriberNSQImpl) Run(url string) {

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
