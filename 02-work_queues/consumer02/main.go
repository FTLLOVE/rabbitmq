package main

import (
	"github.com/streadway/amqp"
	"log"
	"time"
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

	msg, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	failOnError(err, "receiver msg fail.")

	forever := make(chan bool)

	go func() {
		for d := range msg {
			time.Sleep(time.Second * 1)
			log.Printf("consumer02 reveiver a message: %s \n", string(d.Body))
		}
	}()

	log.Println("[*] Waiting for messages. To exit press CTRL+C")

	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}
