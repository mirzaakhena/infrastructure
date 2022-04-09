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

//const exchangeName = "simple.exchange"
//const exchangeType = amqp.ExchangeTopic

var exchangeName = "delayed.exchange"
var exchangeType = "x-delayed-message"

type publisherImpl struct {
	rabbitMQChannel *amqp.Channel
}

// NewPublisher is
// url "amqp://guest:guest@localhost:5672/"
func NewPublisher(url string) *publisherImpl {

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

	return &publisherImpl{
		rabbitMQChannel: ch,
	}
}

// Publish is
func (m *publisherImpl) Publish(topic string, delayInMS int, data payload.Payload) error {

	dataInBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	headers := amqp.Table{
		"x-delay": delayInMS, // only for x-delay-message
	}

	err = m.rabbitMQChannel.Publish(
		exchangeName, // exchange
		topic,        // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        dataInBytes,
			Headers:     headers,
		})
	if err != nil {
		return err
	}

	return nil
}

type subscriberImpl struct {
	queueName string
	topicMap  map[string]HandleFunc
}

// NewSubscriber is
func NewSubscriber(queueName string) Subscriber {
	return &subscriberImpl{
		queueName: queueName,
		topicMap:  map[string]HandleFunc{},
	}
}

func (r *subscriberImpl) Handle(topic string, onReceived HandleFunc) {

	r.topicMap[topic] = onReceived

}

// Run is
// "amqp://guest:guest@localhost:5672/"
func (r *subscriberImpl) Run(url string) {

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

	args := amqp.Table{
		"x-delayed-type": "topic", // only for x-delay-message
	}

	err = rabbitMQChannel.ExchangeDeclare(
		exchangeName, // name
		exchangeType, // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		args,         // arguments
	)
	if err != nil {
		panic(err.Error())
	}

	for s := range r.topicMap {

		q, err := rabbitMQChannel.QueueDeclare(
			r.queueName+"-"+s, // name
			false,             // durable
			false,             // delete when unused
			false,             // exclusive
			false,             // no-wait
			nil,               // arguments
		)
		if err != nil {
			panic(err.Error())
		}

		err = rabbitMQChannel.QueueBind(
			q.Name,       // queue name
			s,            // routing key
			exchangeName, // exchange
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
				//log.Printf("recv %s %s", d.RoutingKey, data.Data)
			}
		}(s)
	}

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)
	<-termChan

}

// https://programmer.ink/think/golang-implements-the-delay-queue-of-rabbitmq.html
// https://stackoverflow.com/questions/52819237/how-to-add-plugin-to-rabbitmq-docker-image
// https://github.com/rabbitmq/rabbitmq-delayed-message-exchange
// https://www.cloudamqp.com/blog/what-is-a-delayed-message-exchange-in-rabbitmq.html
// https://blog.rabbitmq.com/posts/2015/04/scheduling-messages-with-rabbitmq
