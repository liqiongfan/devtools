package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"

	"devtools/icmp"
	"devtools/kafka/consumer"
	"devtools/kafka/producer"
)

func TestPing() {
	v := icmp.Ping(
		icmp.WithDomain("www.baidu.com"),
		icmp.WithCount(4),
		icmp.WithPs(icmp.MacPs),
		icmp.WithEcho(true),
	)
	if v >= 80 {
		fmt.Printf("{%f}", v)
	}
}

func WriteProducer() {
	beginTime := time.Now()
	once := sync.WaitGroup{}
	for x := 0; x < 5; x++ {
		once.Add(1)
		go func(x int) {
			client := producer.NewProducer(producer.WithBootstrapServer([]string{"172.16.119.211:9092"}))
			for i := 0; i < 3000000; i++ {
				_ = producer.PushMessage(
					producer.WithProducer(client),
					producer.WithBuffer([]byte(fmt.Sprintf("Hello lady data: [%d:%d]", x, i))),
					producer.WithTopic("producer_golang_demo_test"),
					producer.WithPartition(int32(x)),
				)
				if i%10000 > 9000 {
					client.Flush(5)
				}
			}
			once.Done()
			client.Flush(50)
			client.Close()
		}(x)
	}
	once.Wait()

	fmt.Printf("Yes: time: %06fs\n", time.Since(beginTime).Seconds())
}

func GetMessages() {
	c := consumer.NewConsumer(
		consumer.WithBootstrapServer([]string{"172.16.119.211:9092"}),
		consumer.WithGroupName("producer_golang_demo_test"),
	)

	v := sync.WaitGroup{}
	for x := 0; x < 20; x++ {
		v.Add(1)
		go func() {
			consumer.GetMessages(
				consumer.WithConsumer(c),
				consumer.WithTopics([]string{"producer_golang_demo_test"}),
				consumer.WithCallback(func(message *kafka.Message, cx chan int) bool {
					fmt.Printf("%s\n", message.Value)
					return true
				}),
			)
			v.Done()
		}()
	}
	v.Wait()
}

func Test() {
	fmt.Printf("begin_time: %d\n", time.Now().Unix())
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGKILL)
	for x := 0; x < 5; x++ {
		go func() {
			GetMessages()
		}()
	}
	<-c
	fmt.Printf("end_time: %d\n", time.Now().Unix())
}

func main() {
	// WriteProducer()
	TestPing()
}
