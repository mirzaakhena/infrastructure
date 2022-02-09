package messaging

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"infrastructure/shared/model/payload"
	"os"
	"os/signal"
	"syscall"
)

const defaultExchange = "TopicChannelExchange"
const exchangeType = "topic"

type publisherRabbitMQImpl struct {
	rabbitMQChannel *amqp.Channel
}

// NewPublisherRabbitMQ is
// url "amqp://guest:guest@localhost:5672/"
func NewPublisherRabbitMQ(url string) *publisherRabbitMQImpl {

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil
	}
	//defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return nil
	}
	//defer ch.Close()

	return &publisherRabbitMQImpl{
		rabbitMQChannel: ch,
	}
}

// Publish is
func (m *publisherRabbitMQImpl) Publish(topic string, data payload.Payload) error {

	dataInBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	err = m.rabbitMQChannel.Publish(
		defaultExchange, // exchange
		topic,           // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        dataInBytes,
		})
	if err != nil {
		return err
	}

	return nil
}

type subscriberRabbitMQImpl struct {
	queueName string
	topicMap  map[string]HandleFunc
}

// NewSubscriberRabbitMQ is
func NewSubscriberRabbitMQ(queueName string) Subscriber {
	return &subscriberRabbitMQImpl{
		queueName: queueName,
		topicMap:  map[string]HandleFunc{},
	}
}

func (r *subscriberRabbitMQImpl) Handle(topic string, onReceived HandleFunc) {

	r.topicMap[topic] = onReceived

}

// Run is
// "amqp://guest:guest@localhost:5672/"
func (r *subscriberRabbitMQImpl) Run(url string) {

	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err.Error())
	}
	defer func(conn *amqp.Connection) {
		err := conn.Close()
		if err != nil {
			panic(err.Error())
		}
	}(conn)

	rabbitMQChannel, err := conn.Channel()
	if err != nil {
		panic(err.Error())
	}
	defer func() {
		err := rabbitMQChannel.Close()
		if err != nil {
			panic(err.Error())
		}
	}()

	err = rabbitMQChannel.ExchangeDeclare(
		defaultExchange, // name
		exchangeType,    // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		panic(err.Error())
	}

	for s := range r.topicMap {

		q, err := rabbitMQChannel.QueueDeclare(
			fmt.Sprintf("%s-%s", r.queueName, s), // name
			false,                                // durable
			false,                                // delete when unused
			false,                                // exclusive
			false,                                // no-wait
			nil,                                  // arguments
		)
		if err != nil {
			panic(err.Error())
		}

		err = rabbitMQChannel.QueueBind(
			q.Name,          // queue name
			s,               // routing key
			defaultExchange, // exchange
			false,
			nil,
		)
		if err != nil {
			panic(err.Error())
		}

		deliveryMsg, err := rabbitMQChannel.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("%s %s\n", q.Name, s)

		go func(routingKey string) {
			for d := range deliveryMsg {
				var data payload.Payload
				err := json.Unmarshal(d.Body, &data)
				r.topicMap[routingKey](data, err)
			}
		}(s)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

}
