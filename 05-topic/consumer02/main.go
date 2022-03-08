package main

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	connection, err := amqp.Dial("amqp://admin:admin@localhost:5672")
	failOnError(err, "connect rabbitmq server fail.")
	defer connection.Close()

	channel, err := connection.Channel()
	failOnError(err, "open channel fail.")
	defer channel.Close()

	var exchangeName = "exchange_topic"

	var queueName01 = "queue_topic_01"
	var queueName02 = "queue_topic_02"
	var queueName03 = "queue_topic_03"

	channel.QueueDeclare(queueName01, true, false, false, false, nil)
	channel.QueueDeclare(queueName02, true, false, false, false, nil)
	channel.QueueDeclare(queueName03, true, false, false, false, nil)

	channel.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil)

	channel.QueueBind(queueName01, "item.#.hello", exchangeName, false, nil)
	channel.QueueBind(queueName02, "item.*.*", exchangeName, false, nil)
	channel.QueueBind(queueName03, "*.*.*", exchangeName, false, nil)

	msg, _ := channel.Consume(queueName02, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("consumer02(item.*.*) receiver msg: %s \n", d.Body)
		}
	}()

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}
