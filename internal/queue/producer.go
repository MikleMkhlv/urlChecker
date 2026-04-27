package queue

import "github.com/rabbitmq/amqp091-go"

type Producer struct {
	ch        *amqp091.Channel
	queueName string
}

func (p *Producer) PublishMessage(body []byte) error {
	return p.ch.Publish(
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers: amqp091.Table{
				"x-retry": int32(0),
			},
		},
	)
}
