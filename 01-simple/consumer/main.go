package main

import (
	"github.com/streadway/amqp"
	"log"
)

func main() {
	// 声明连接(connection)
	connection, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	failOnError(err, "connect rabbitmq server fail.")
	defer connection.Close()

	// 从连接中声明channel
	channel, err := connection.Channel()
	failOnError(err, "open channel fail.")
	defer channel.Close()

	// 从channel中声明队列(queue)
	queue, err := channel.QueueDeclare("01-simple", false, false, false, false, nil)
	failOnError(err, "declare queue fail.")

	// 从默认的交换机中的队列获取消息
	msg, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	failOnError(err, "receiver msg fail.")

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("reveiver a message: %s \n", string(d.Body))
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
