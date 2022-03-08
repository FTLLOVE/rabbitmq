package rabbitmq

import (
	"errors"
	"github.com/streadway/amqp"
	"log"
)

// RabbitMQ 定义rabbitmq结构体
type RabbitMQ struct {
	conn         *amqp.Connection // 连接实例
	channel      *amqp.Channel    // 管道
	ExchangeName string           // 交换机
	QueueName    string           // 队列
	Key          string           // routing key/binding key
}

// RabbitMQInterface 定义rabbitmq的生产者和消费者的接口
type RabbitMQInterface interface {
	Publish(message string) error
	Consume() (<-chan amqp.Delivery, error)
}

// NewRabbitMQ 定义rabbitmq的构造方法
func NewRabbitMQ(exchangeName, queueName, key string) (*RabbitMQ, error) {
	var err error

	rabbitMQ := &RabbitMQ{
		ExchangeName: exchangeName,
		QueueName:    queueName,
		Key:          key,
	}

	rabbitMQ.conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("amqp dial connect fail,message: %s \n", err.Error())
		return nil, err
	}

	rabbitMQ.channel, err = rabbitMQ.conn.Channel()
	if err != nil {
		log.Fatalf("create channel fail,message: %s \n", err.Error())
		return nil, err
	}

	return rabbitMQ, nil
}

// Close 关闭连接资源
func (this *RabbitMQ) Close() {
	this.channel.Close()
	this.conn.Close()
}

// Simple 简单模式的结构体
type Simple struct {
	*RabbitMQ
}

// NewRabbitMQSimple 简单模式下的实例构建
// 简单模式下只需要队列即可，
// 使用默认的交换机
func NewRabbitMQSimple(queueName string) (*Simple, error) {
	if queueName == "" {
		log.Fatalf("queueName is required of simple mode,queueName=%s  \n", queueName)
		return nil, errors.New("queueName is required of simple mode")
	}
	rabbitMQ, err := NewRabbitMQ("", queueName, "")
	if err != nil {
		return nil, err
	}
	return &Simple{
		RabbitMQ: rabbitMQ,
	}, nil
}

// Publish 简单模式下的生产者
// 简单模式下只需要队列，
// 使用默认的交换机
func (this *Simple) Publish(message string) error {
	var (
		err   error
		queue amqp.Queue
	)
	// 声明队列
	if queue, err = this.channel.QueueDeclare(
		this.QueueName,
		true,
		false,
		false,
		false,
		nil); err != nil {
		log.Fatalf("queue declare fail of simple mode,message: %s \n", err.Error())
		return err
	}

	// 发布消息
	if err = this.channel.Publish(
		this.ExchangeName,
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	); err != nil {
		log.Fatalf("publish message fail of simple model,message: %s \n", err.Error())
		return err
	}

	return nil
}

// Consume 简单模式下的消费者
func (this *Simple) Consume() (<-chan amqp.Delivery, error) {

	var err error

	// 声明队列
	if _, err = this.channel.QueueDeclare(
		this.QueueName,
		true,
		false,
		false,
		false,
		nil); err != nil {
		log.Fatalf("queue declare fail of simple mode,message: %s \n", err.Error())
		return nil, err
	}

	// 消费
	message, err := this.channel.Consume(
		this.ExchangeName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume message fail of simple mode,message: %s \n", err.Error())
		return nil, err
	}

	return message, nil
}

// PubSub 发布订阅模式下的结构体
type PubSub struct {
	*RabbitMQ
}

// NewRabbitMQPubSub  发布订阅模式下的实例构建
// 发布订阅模式下需要传递交换机，并且需要建立绑定关系
// 并且交换机的的类型必须是"fanout"
func NewRabbitMQPubSub(exchangeName, queueName string) (*PubSub, error) {
	if exchangeName == "" {
		log.Fatalf("exchangeName  is required of pub/sub mode,exchangeName=%s\n", exchangeName)
		return nil, errors.New("exchangeName is required of pub/sub mode")
	}
	rabbitMQ, err := NewRabbitMQ(exchangeName, queueName, "")
	if err != nil {
		return nil, err
	}
	return &PubSub{
		RabbitMQ: rabbitMQ,
	}, nil
}

// Publish 发布订阅模式下的生产者
func (this *PubSub) Publish(message string) error {

	var (
		err error
	)

	// 声明交换机，交换机类型必须是"fanout"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of pub/sub mode,message: %s \n", err.Error())
		return err
	}

	// 发布消息
	if err = this.channel.Publish(
		this.ExchangeName,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	); err != nil {
		log.Fatalf("publish message fail of pub/sub model,message: %s \n", err.Error())
		return err
	}

	return nil
}

// Consume 发布订阅模式下的消费者
func (this *PubSub) Consume() (<-chan amqp.Delivery, error) {

	var (
		err   error
		queue amqp.Queue
	)

	// 声明队列
	if queue, err = this.channel.QueueDeclare(
		this.QueueName,
		true,
		false,
		true,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue declare fail of pub/sub mode,message: %s \n", err.Error())
		return nil, err
	}

	// 声明交换机，交换机类型必须是"fanout"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeFanout,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of pub/sub mode,message: %s \n", err.Error())
		return nil, err
	}

	// 交换机和队列进行绑定,routing key为空
	if err = this.channel.QueueBind(
		queue.Name,
		"",
		this.ExchangeName,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange and queue bind fail of pub/sub mode,message: %s \n", err.Error())
		return nil, err
	}

	// 消费
	message, err := this.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume message fail of pub/sub mode,message: %s \n", err.Error())
		return nil, err
	}

	return message, nil
}

// Routing 路由模式下的结构体
type Routing struct {
	*RabbitMQ
}

// NewRabbitMQRouting 路由模式下的实例构建
// 路由模式下需要交换机，key
// 交换的类型必须是"direct"
func NewRabbitMQRouting(exchangeName, key string) (*Routing, error) {
	if exchangeName == "" || key == "" {
		log.Fatalf("exchangeName and key is required of routing mode,exchangeName=%s ,key=%s \n", exchangeName, key)
		return nil, errors.New("exchangeName and key is required of routing mode")
	}
	rabbitMQ, err := NewRabbitMQ(exchangeName, "", key)
	if err != nil {
		return nil, err
	}

	return &Routing{
		RabbitMQ: rabbitMQ,
	}, nil
}

// Publish 路由模式下的生产者
func (this *Routing) Publish(message string) error {

	var (
		err error
	)

	// 声明交换机，交换机的类型必须是"direct"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of routing mode,message: %s \n", err.Error())
		return err
	}

	// 发布消息
	if err = this.channel.Publish(
		this.ExchangeName,
		this.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	); err != nil {
		log.Fatalf("publish message fail of routing model,message: %s \n", err.Error())
		return err
	}

	return nil
}

// Consume 路由模式下的消费者
func (this *Routing) Consume() (<-chan amqp.Delivery, error) {

	var (
		err   error
		queue amqp.Queue
	)

	// 声明队列
	if queue, err = this.channel.QueueDeclare(
		this.QueueName,
		true,
		false,
		true,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue declare fail of routing mode,message: %s \n", err.Error())
		return nil, err
	}

	// 声明交换机，交换机的类型必须是"direct"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeDirect,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of routing mode,message: %s \n", err.Error())
		return nil, err
	}

	// 交换机和队列进行绑定, key非常重要
	if err = this.channel.QueueBind(
		queue.Name,
		this.Key,
		this.ExchangeName,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange and queue bind fail of routing mode,message: %s \n", err.Error())
		return nil, err
	}

	// 消费消息
	message, err := this.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume message fail of routing mode,message: %s \n", err.Error())
		return nil, err
	}

	return message, nil

}

// Topic 主题模式下的结构体
type Topic struct {
	*RabbitMQ
}

// NewRabbitMQTopic 主题模式下的实例构建
// 主题模式下需要交换机，key
// 主题模式下的key更加灵活，#代表一个或多个词，*代表一个词
func NewRabbitMQTopic(exchangeName, key string) (*Topic, error) {
	if exchangeName == "" || key == "" {
		log.Fatalf("exchangeName and key is required of routing mode,exchangeName=%s, key=%s \n", exchangeName, key)
		return nil, errors.New("exchangeName and key is required of routing mode")
	}
	rabbitMQ, err := NewRabbitMQ(exchangeName, "", key)
	if err != nil {
		return nil, err
	}

	return &Topic{
		RabbitMQ: rabbitMQ,
	}, nil
}

// Publish 主题模式下的生产者
func (this *Topic) Publish(message string) error {

	var (
		err error
	)

	// 声明交换机, 交换机的类型必须是"topic"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of topic mode,message: %s \n", err.Error())
		return err
	}

	// 发布消息
	if err = this.channel.Publish(
		this.ExchangeName,
		this.Key,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		},
	); err != nil {
		log.Fatalf("publish message fail of topic model,message: %s \n", err.Error())
		return err
	}

	return nil
}

// Consume 主题模式下的消费者
func (this *Topic) Consume() (<-chan amqp.Delivery, error) {

	var (
		err   error
		queue amqp.Queue
	)

	// 声明队列
	if queue, err = this.channel.QueueDeclare(
		this.QueueName,
		true,
		false,
		true,
		false,
		nil,
	); err != nil {
		log.Fatalf("queue declare fail of topic mode,message: %s \n", err.Error())
		return nil, err
	}

	// 声明交换机, 交换机的类型必须是"topic"
	if err = this.channel.ExchangeDeclare(
		this.ExchangeName,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchage declare fail of topic mode,message: %s \n", err.Error())
		return nil, err
	}

	// 交换机和队列进行绑定，key非常的重要，且灵活
	if err = this.channel.QueueBind(
		queue.Name,
		this.Key,
		this.ExchangeName,
		false,
		nil,
	); err != nil {
		log.Fatalf("exchange and queue bind fail of topic mode,message: %s \n", err.Error())
		return nil, err
	}

	message, err := this.channel.Consume(
		queue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("consume message fail of topic mode,message: %s \n", err.Error())
		return nil, err
	}

	return message, nil
}
