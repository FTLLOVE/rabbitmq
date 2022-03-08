package main

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func main() {
	connection, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	failOnError(err, "connect rabbitmq server fail.")
	defer connection.Close()

	channel, err := connection.Channel()
	failOnError(err, "open channel fail.")
	defer channel.Close()

	queue, err := channel.QueueDeclare("02-work_queues", false, false, false, false, nil)
	failOnError(err, "declare queue fail.")

	for i := 0; i < 200; i++ {
		msg := fmt.Sprintf("hello world %d", i+1)
		err = channel.Publish("", queue.Name, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		failOnError(err, "publish a message fail.")
	}

}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}
