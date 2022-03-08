# rabbitmq 消息队列

## 1. 简单模式(simple mode)

> 简单模式下是不需要`交换机（exchange）`的，只需要队列即可，但是不使用交换机不代表没有使用交换机，而是使用rabbitmq提供的默认的交换机(default)。```一个生产者对应一个消费者```

> 生产者代码如下:

```go
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

```

> 消费者代码如下:

```go
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

```

## 2. 工作队列模式(work_queue)

> 工作队列模式相比简单模式下，多了一个或者多个消费者，多个消费者共同消费同一个队列中的消息。一般在任务多或者任务重的情景下使用工作模式可以提高工作效率。是一种抢占式的消费。```一个生产者对应多个消费者```

> 生产者代码如下:

```go
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
```

> 消费者代码如下:

```go
package main

import (
	"github.com/streadway/amqp"
	"log"
	"time"
)

func main() {
	connection, err := amqp.Dial("amqp://admin:admin@localhost:5672")
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
			log.Printf("consumer01 reveiver a message: %s \n", string(d.Body))
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

```

## 3. 发布订阅模式(pub/sub)

> 发布订阅相比较之前的模式而言，多了一个`交换机`。每个消费者监听他们自己的队列即可完成消费。
> 交换机有多种类型: topic的类型可以做到`direct`和`fanout`的效果。其中headers不是特别的常用，不做讲解。

```
Fanout：广播，将消息交给所有绑定到交换机的队列 (只要队列和交换机绑定，这个时候routing key也无作用，所有的队列都能收到消息)
Direct：定向，把消息交给符合指定routing key 的队列(一对一绑定)
Topic：通配符，把消息交给符合routing pattern（路由模式） 的队列(支持通配符绑定)
```

> 生产者代码如下:

```go
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

```

> 消费者1代码如下:

```go
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

```

> 消费者2代码如下:

```go
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
	var queueName = "queue_console"

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
			log.Printf("控制台服务器接收到了消息: %s \n", string(d.Body))
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

```

## 4. 路由模式(routing)

> 路由模式下，队列和交换机不能随意绑定，必须通过routing key进行绑定，生产者发送消息到exchange，也需要指定消息的routing key, 交换机发送消息需要根据routing key进行发送给队列，而且必须完全一致才可以发送。

> 生产者代码如下:

```go
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

	var msg = "插入了一条数据"
	_ = channel.Publish(exchangeName, "insert", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}

```

> 消费者1代码如下:

```go
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

```

> 消费者2代码如下:

```go
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

	msg, _ := channel.Consume(queueName02, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("consumer02(insert/update/delete) update got message: %s \n", string(d.Body))
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

```

## 5. 通配符模式(topic)

> 通配符模式下，交换机和队列可以绑定的更加灵活 <br>
> '#' 代表一个字符 <br>
> '*' 代表一个或者多个字符

> 生产者代码如下:

```go
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

	var msg = "我是发送过去的数据哈哈"

	channel.Publish(exchangeName, "123.456.789", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("msg: %s, err: %s \n", msg, err.Error())
		return
	}
}

```

> 消费者1代码如下:

```go
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

	msg, _ := channel.Consume(queueName01, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("consumer01(item.#.hello) receiver msg: %s \n", d.Body)
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

```

> 消费者2代码如下:

```go
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

```

> 消费者3代码如下:

```go
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

	msg, _ := channel.Consume(queueName03, "", true, false, false, false, nil)

	forever := make(chan bool)

	go func() {
		for d := range msg {
			log.Printf("consumer03(*.*.*) receiver msg: %s \n", d.Body)
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

```