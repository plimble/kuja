package rabbitmq

import (
	"github.com/plimble/kuja/broker"
	"github.com/streadway/amqp"
)

type rabbitmqBroker struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
}

func NewBroker(url string) *rabbitmqBroker {
	return &rabbitmqBroker{
		url: url,
	}
}

func (r *rabbitmqBroker) Connect() error {
	var err error
	r.conn, err = amqp.Dial(r.url)
	if err != nil {
		return err
	}

	r.ch, err = r.conn.Channel()
	if err != nil {
		return err
	}

	return nil
}

func (r *rabbitmqBroker) Close() {
	r.ch.Close()
	r.conn.Close()
}

func (r *rabbitmqBroker) Publish(topic string, msg *broker.Message) error {
	r.ch.ExchangeDeclare(
		topic,    // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	data, err := msg.Marshal()
	if err != nil {
		return err
	}

	return r.ch.Publish(
		topic, // exchange
		"",    // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			// ContentType: "text/plain",
			Body: data,
		},
	)
}

func (r *rabbitmqBroker) Subscribe(topic, queue, appId string, h broker.Handler) {
	r.ch.ExchangeDeclare(
		topic,    // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)

	q, _ := r.ch.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	r.ch.QueueBind(
		q.Name, // queue name
		"",     // routing key
		topic,  // exchange
		false,
		nil,
	)

	consume, _ := r.ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // no-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)

	go func() {
		for d := range consume {
			brokerMsg := &broker.Message{}
			brokerMsg.Unmarshal(d.Body)
			retryCount, err := h(topic, brokerMsg)
			if err == nil {
				d.Ack(false)
			} else {
				for i := 0; i < retryCount; i++ {
					brokerMsg.Retry++
					_, err := h(topic, brokerMsg)
					if err == nil {
						d.Ack(false)
						break
					}
				}
				d.Reject(false)
			}
		}
	}()
}
