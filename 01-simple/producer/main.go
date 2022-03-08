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

	// channel发送消息到默认的交换机,然后由默认的交换机发送到指定的队列(queue)
	var msg = "hello world"
	err = channel.Publish("", queue.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
	failOnError(err, "publish a message fail.")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}
