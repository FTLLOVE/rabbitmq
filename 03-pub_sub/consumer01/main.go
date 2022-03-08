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

	var exchangeName = "exchange_pub_sub"
	var queueName = "queue_email"

	queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
	failOnError(err, "declare queue fail.")

	err = channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "declare exchange fail.")

	// 交换机和队列绑定
	err = channel.QueueBind(queue.Name, "", exchangeName, false, nil)
	failOnError(err, "exchange and queue bind fail.")

	msg, err := channel.Consume(queue.Name, "", true, false, false, false, nil)
	failOnError(err, "receiver msg fail.")

	forever := make(chan bool)

	go func() {
		for d := range msg {
			time.Sleep(time.Second * 1)
			log.Printf("邮件服务器接收到了消息: %s \n", string(d.Body))
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
