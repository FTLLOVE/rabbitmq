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

	var exchangeName = "exchange_routing"

	var queueName01 = "queue_01"
	var queueName02 = "queue_02"

	channel.QueueDeclare(queueName01, true, false, false, false, nil)
	channel.QueueDeclare(queueName02, true, false, false, false, nil)

	channel.ExchangeDeclare(exchangeName, "direct", true, false, false, false, nil)

	channel.QueueBind(queueName01, "insert", exchangeName, false, nil)

	channel.QueueBind(queueName02, "insert", exchangeName, false, nil)
	channel.QueueBind(queueName02, "update", exchangeName, false, nil)
	channel.QueueBind(queueName02, "delete", exchangeName, false, nil)

	msg, _ := channel.Consume(queueName01, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("consumer01(insert) insert got message: %s \n", string(d.Body))
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
