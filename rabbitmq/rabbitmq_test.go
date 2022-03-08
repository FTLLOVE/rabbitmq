package rabbitmq

import (
	"strconv"
	"testing"
)

var (
	simpleQueueName = "queue_rabbitmq_simple_demo01"

	pubSubExchangeName = "exchange_pub_sub_demo"
	pubSubQueueName01  = "queue_rabbitmq_pub_sub_01"
	pubSubQueueName02  = "queue_rabbitmq_pub_sub_02"

	routingExchangeName = "exchange_routing_demo"
	routingQueueName01  = "queue_rabbitmq_routing_demo_01"
	routingQueueName02  = "queue_rabbitmq_routing_demo_02"

	topicExchangeName = "exchange_topic_demo"
	topicQueueName01  = "queue_rabbitmq_topic_demo_01"
	topicQueueName02  = "queue_rabbitmq_topic_demo_02"
	topicQueueName03  = "queue_rabbitmq_topic_demo_03"
)

func TestSimple_Publish(t *testing.T) {
	mq, _ := NewRabbitMQSimple(simpleQueueName)
	defer mq.Close()

	for i := 0; i < 10; i++ {
		mq.Publish("simple-" + strconv.Itoa(i+1))
	}
}

func TestSimple_Consume(t *testing.T) {
	mq, _ := NewRabbitMQSimple(simpleQueueName)

	forever := make(chan bool)
	go func() {
		message, _ := mq.Consume()
		for d := range message {
			t.Log("simple-" + string(d.Body))
		}
	}()

	<-forever
}

func TestWorkQueue_Consume(t *testing.T) {
	mq1, _ := NewRabbitMQSimple(simpleQueueName)
	mq2, _ := NewRabbitMQSimple(simpleQueueName)

	forever := make(chan bool)

	go func() {
		message, _ := mq1.Consume()
		for d := range message {
			t.Log("work_queue_consume01-" + string(d.Body))
		}
	}()

	go func() {
		message, _ := mq2.Consume()
		for d := range message {
			t.Log("work_queue_consume02-" + string(d.Body))
		}
	}()
	<-forever
}

// 注意：发布订阅是一种广播的机制，只要队列和交换机绑定好关系即可
// 发送消息是发送到交换机，由交换机根据类型来发送到队列中,和发送方没什么关系
func TestPubSub_Publish(t *testing.T) {
	// 这边队列名称1起不到任何作用
	mq, _ := NewRabbitMQPubSub(pubSubExchangeName, "")
	defer mq.Close()

	for i := 0; i < 10; i++ {
		mq.Publish("pub/sub-message-" + strconv.Itoa(i+1))
	}
}

func TestPubSub_Consume(t *testing.T) {
	mq1, _ := NewRabbitMQPubSub(pubSubExchangeName, pubSubQueueName01)
	defer mq1.Close()

	mq2, _ := NewRabbitMQPubSub(pubSubExchangeName, pubSubQueueName02)
	defer mq2.Close()

	forever := make(chan bool)

	go func() {
		message, _ := mq1.Consume()
		for d := range message {
			t.Log("pub/sub_consume01-" + string(d.Body))
		}
	}()

	go func() {
		message, _ := mq2.Consume()
		for d := range message {
			t.Log("pub/sub_consume02-" + string(d.Body))
		}
	}()

	<-forever
}

func TestRouting_Publish(t *testing.T) {
	mq, _ := NewRabbitMQRouting(routingExchangeName, "key.one")
	defer mq.Close()
	for i := 0; i < 10; i++ {
		mq.Publish("routing-message-" + strconv.Itoa(i+1))
	}
}

func TestRouting_Consume(t *testing.T) {
	mq1, _ := NewRabbitMQRouting(routingExchangeName, "key.one")
	mq2, _ := NewRabbitMQRouting(routingExchangeName, "key.one")
	mq3, _ := NewRabbitMQRouting(routingExchangeName, "key.two")

	defer mq1.Close()
	defer mq2.Close()
	defer mq3.Close()

	forever := make(chan bool)

	go func() {
		message, _ := mq1.Consume()
		for d := range message {
			t.Log("routing-consumer01(key.one)-" + string(d.Body))
		}
	}()

	go func() {
		message, _ := mq2.Consume()
		for d := range message {
			t.Log("routing-consumer02(key.one/key.two)-" + string(d.Body))
		}
	}()

	go func() {
		message, _ := mq3.Consume()
		for d := range message {
			t.Log("routing-consumer02(key.one/key.two)-" + string(d.Body))

		}
	}()

	<-forever
}

func TestTopic_Publish(t *testing.T) {
	mq, _ := NewRabbitMQTopic(topicExchangeName, "key.two")

	defer mq.Close()

	for i := 0; i < 10; i++ {
		mq.Publish("topic-message-" + strconv.Itoa(i+1))
	}
}

func TestTopic_Consume(t *testing.T) {
	mq1, _ := NewRabbitMQTopic(topicExchangeName, "key.one")
	mq2, _ := NewRabbitMQTopic(topicExchangeName, "key.*")
	mq3, _ := NewRabbitMQTopic(topicExchangeName, "key.#")

	defer mq1.Close()
	defer mq2.Close()
	defer mq3.Close()

	forever := make(chan bool)

	go func() {
		message, _ := mq1.Consume()
		for d := range message {
			t.Log("topic-consumer1(key.one)-" + string(d.Body))

		}
	}()

	go func() {
		message, _ := mq2.Consume()
		for d := range message {
			t.Log("topic-consumer2(key.#)-" + string(d.Body))

		}
	}()

	go func() {
		message, _ := mq3.Consume()
		for d := range message {
			t.Log("topic-consumer3(key.#)-" + string(d.Body))

		}
	}()

	<-forever
}
