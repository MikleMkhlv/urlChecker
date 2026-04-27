package queue

import (
	"encoding/json"
	"fmt"
	"watchdog/internal/storage"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch        *amqp091.Channel
	queueName string
	store     *storage.Storage
}

func (c *Consumer) Consume(workerId int, handler func(int, []byte) error) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		var st storage.Sites
		json.Unmarshal(msg.Body, &st)
		retry, ok := msg.Headers["x-retry"].(int32)
		if !ok {
			retry = 0
		}
		if err := handler(workerId, msg.Body); err != nil {
			if st.IsDQL {
				fmt.Printf("url=%s is DLQ queue\n", st.Url)
				msg.Ack(false)
				continue
			} else if retry >= 3 {
				st.IsDQL = true
				newBody, err := json.Marshal(st)
				if err != nil {
					fmt.Printf("error marshal site url: %s\n", st.Url)
					msg.Ack(false)
					continue
				}
				msg.Body = newBody
				err = c.store.UpdateDLQStatus(st, true)
				if err != nil {
					fmt.Printf("Error update dlq status is database for url=%s\n", st.Url)
					msg.Ack(false)
					continue
				}
				err = publishInDLQ(c, msg)
				if err != nil {
					fmt.Printf("Error send url=%s if DLQ queue\n", st.Url)
					msg.Ack(false)
					continue
				}
				fmt.Printf("[Worker-%d]: url=%s sent in DLQ queue\n", workerId, st.Url)
			} else {
				fmt.Printf("[Worker-%d]: url=%s retryed\n", workerId, st.Url)
				publishToRetryQueue(c, msg, retry)
			}
		}
		msg.Ack(false)
	}

	return nil
}

// func repuiblish(consumer *Consumer, msg amqp091.Delivery, countRetry int32) {
// 	consumer.ch.Publish(
// 		"",
// 		consumer.queueName,
// 		false,
// 		false,
// 		amqp091.Publishing{
// 			ContentType: "application/json",
// 			Body:        msg.Body,
// 			Headers: amqp091.Table{
// 				"x-retry": countRetry + 1,
// 			},
// 		},
// 	)
// 	msg.Ack(false)
// }

func publishToRetryQueue(consumer *Consumer, msg amqp091.Delivery, countRetry int32) error {
	dlqQueue := "retry_queue"
	_, err := consumer.ch.QueueDeclare(
		dlqQueue,
		true,
		false,
		false,
		false,
		amqp091.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": "sites_queue",
		},
	)
	if err != nil {
		return err
	}

	consumer.ch.Publish(
		"",
		dlqQueue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
			Headers: amqp091.Table{
				"x-retry": countRetry + 1,
			},
			Expiration: "10000",
		},
	)
	return nil
}

func publishInDLQ(consumer *Consumer, msg amqp091.Delivery) error {
	dlqQueue := "failed_queue"
	_, err := consumer.ch.QueueDeclare(
		dlqQueue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	consumer.ch.Publish(
		"",
		dlqQueue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        msg.Body,
		},
	)
	return nil
}
