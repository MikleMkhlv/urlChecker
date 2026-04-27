package queue

import (
	"watchdog/internal/storage"

	"github.com/rabbitmq/amqp091-go"
)

type Rabbit struct {
	conn  *amqp091.Connection
	store *storage.Storage
}

func New(url string, store *storage.Storage) (*Rabbit, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, err
	}
	return &Rabbit{
		conn:  conn,
		store: store,
	}, nil
}

func (mq *Rabbit) NewProducer(queueName string) (*Producer, error) {
	ch, err := mq.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = mq.createQueue(ch, queueName)
	if err != nil {
		ch.Close()
		return nil, err
	}

	return &Producer{
		ch:        ch,
		queueName: queueName,
	}, nil
}

func (mq *Rabbit) createQueue(ch *amqp091.Channel, queueName string) error {
	_, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	return nil
}

func (mq *Rabbit) NewConsumer(queueName string) (*Consumer, error) {
	ch, err := mq.conn.Channel()
	if err != nil {
		return nil, err
	}

	err = mq.createQueue(ch, queueName)
	if err != nil {
		ch.Close()
		return nil, err
	}
	err = ch.Qos(1, 0, false)

	return &Consumer{
		ch:        ch,
		queueName: queueName,
		store:     mq.store,
	}, nil
}

func (mq *Rabbit) Close() error {
	return mq.conn.Close()
}
