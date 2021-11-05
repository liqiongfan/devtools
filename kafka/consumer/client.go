package consumer

import (
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type newOption struct {
	bootstrapServer          []string
	groupId, autoOffsetReset string
	otherOptions             map[string]interface{}
}

type NewOptions func(*newOption)

func WithBootstrapServer(bootstrapServer []string) NewOptions {
	return func(o *newOption) {
		o.bootstrapServer = bootstrapServer
	}
}

func WithOtherOptions(otherOptions map[string]interface{}) NewOptions {
	return func(o *newOption) {
		o.otherOptions = otherOptions
	}
}

func WithGroupName(groupId string) NewOptions {
	return func(o *newOption) {
		o.groupId = groupId
	}
}

func WithAutoOffsetReset(autoOffsetReset string) NewOptions {
	return func(o *newOption) {
		o.autoOffsetReset = autoOffsetReset
	}
}

func NewConsumer(opts ...NewOptions) *kafka.Consumer {
	opt := &newOption{
		bootstrapServer: []string{"localhost"},
		groupId:         "",
		autoOffsetReset: "beginning",
		otherOptions:    nil,
	}
	for _, o := range opts {
		o(opt)
	}

	configMap := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(opt.bootstrapServer, ","),
		"group.id":          opt.groupId,
		"auto.offset.reset": opt.autoOffsetReset,
	}

	if opt.otherOptions != nil && len(opt.otherOptions) > 0 {
		for key, value := range opt.otherOptions {
			err := configMap.SetKey(key, value)
			if err != nil {
				panic(err)
			}
		}
	}
	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		panic(err)
	}

	return consumer
}

type getOption struct {
	consumer       *kafka.Consumer
	topics         []string
	reBalance      kafka.RebalanceCb
	callback       func(*kafka.Message, chan int) bool
	goroutineCount int
}
type GetOptions func(*getOption)

func WithConsumer(c *kafka.Consumer) GetOptions {
	return func(o *getOption) {
		o.consumer = c
	}
}

func WithTopics(topics []string) GetOptions {
	return func(o *getOption) {
		o.topics = topics
	}
}

func WithReBalanceCB(reBalanceCb kafka.RebalanceCb) GetOptions {
	return func(o *getOption) {
		o.reBalance = reBalanceCb
	}
}

func WithCallback(cb func(*kafka.Message, chan int) bool) GetOptions {
	return func(o *getOption) {
		o.callback = cb
	}
}

func WithGoroutineCount(c int) GetOptions {
	return func(o *getOption) {
		o.goroutineCount = c
	}
}

func GetMessages(opts ...GetOptions) {
	opt := &getOption{
		consumer:       nil,
		topics:         []string{"*"}, // Listen all topics
		reBalance:      nil,
		callback:       nil,
		goroutineCount: 2000, // default goroutine was 2000
	}
	for _, o := range opts {
		o(opt)
	}

	if opt.consumer == nil {
		panic("kafka consumer should be initialised")
	}

	if opt.callback == nil {
		panic("consumer message should be callback")
	}

	c := opt.consumer

	err := c.SubscribeTopics(opt.topics, opt.reBalance)
	if err != nil {
		panic(err)
	}

	cx := make(chan int, opt.goroutineCount)

	for {
		message, err := c.ReadMessage(-1)
		if err != nil {
			panic(err)
		}
		if opt.callback(message, cx) {
			_, _ = c.CommitMessage(message)
		}
	}
}
