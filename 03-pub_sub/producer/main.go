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

	// 定义交换机(fanout)
	var exchangeName = "exchange_pub_sub"
	err = channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	failOnError(err, "declare exchange fail.")

	for i := 0; i < 20; i++ {
		var msg = fmt.Sprintf("我是一个日志 %d", i+1)
		err = channel.Publish(exchangeName, "", false, false, amqp.Publishing{
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
