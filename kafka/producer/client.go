package producer

import (
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type pushOption struct {
	p          *kafka.Producer
	topic      string
	partition  int32
	data       []byte
	offset     kafka.Offset
	metadata   *string
	key        []byte
	waitResult bool
}
type PushOptions func(option *pushOption)

type newOption struct {
	bootstrapServers                                             []string
	messageMaxBytes, messageCopyMaxBytes, receiveMessageMaxBytes int
	otherOptions                                                 map[string]interface{}
}
type NewOptions func(*newOption)

func WithBootstrapServer(bootstrapServer []string) NewOptions {
	return func(o *newOption) {
		o.bootstrapServers = bootstrapServer
	}
}

func WithMessageMaxBytes(messageMaxBytes int) NewOptions {
	return func(o *newOption) {
		o.messageMaxBytes = messageMaxBytes
	}
}

func WithMessageCopyMaxBytes(messageCopyMaxBytes int) NewOptions {
	return func(o *newOption) {
		o.messageCopyMaxBytes = messageCopyMaxBytes
	}
}

func WithReceiveMessageMaxBytes(receiveMessageMaxBytes int) NewOptions {
	return func(o *newOption) {
		o.receiveMessageMaxBytes = receiveMessageMaxBytes
	}
}

func WithOtherOptions(otherOptions map[string]interface{}) NewOptions {
	return func(o *newOption) {
		o.otherOptions = otherOptions
	}
}

func NewProducer(options ...NewOptions) *kafka.Producer {

	opt := &newOption{
		bootstrapServers:    []string{"localhost"},
		messageMaxBytes:     1000000,
		messageCopyMaxBytes: 65535,
		/*建议值：fetch.message.max.bytes * 分区数 + 消息的最大字节数. */
		receiveMessageMaxBytes: 100000000,
	}
	for _, o := range options {
		o(opt)
	}

	configMap := &kafka.ConfigMap{
		"bootstrap.servers":            strings.Join(opt.bootstrapServers, ","),
		"message.max.bytes":            opt.messageMaxBytes,
		"message.copy.max.bytes":       opt.messageCopyMaxBytes,
		"receive.message.max.bytes":    opt.receiveMessageMaxBytes,
		"queue.buffering.max.messages": 100000,
	}

	if opt.otherOptions != nil && len(opt.otherOptions) > 0 {
		for key, value := range opt.otherOptions {
			err := configMap.SetKey(key, value)
			if err != nil {
				panic(err)
			}
		}
	}

	pp, err := kafka.NewProducer(configMap)
	if err != nil {
		panic(err)
	}
	return pp
}

func WithProducer(p *kafka.Producer) PushOptions {
	return func(o *pushOption) {
		o.p = p
	}
}

func WithTopic(topic string) PushOptions {
	return func(o *pushOption) {
		o.topic = topic
	}
}

func WithPartition(partition int32) PushOptions {
	return func(o *pushOption) {
		o.partition = partition
	}
}

func WithBuffer(data []byte) PushOptions {
	return func(o *pushOption) {
		if len(data) > 0 {
			o.data = data
		}
	}
}

func WithOffset(offset kafka.Offset) PushOptions {
	return func(o *pushOption) {
		o.offset = offset
	}
}

func WithKey(key string) PushOptions {
	return func(o *pushOption) {
		if len(key) > 0 {
			o.key = []byte(key)
		}
	}
}

func WithWaitResult(wait bool) PushOptions {
	return func(o *pushOption) {
		o.waitResult = wait
	}
}

func PushMessage(opts ...PushOptions) bool {
	opt := &pushOption{
		p:          nil,
		topic:      "",
		partition:  kafka.PartitionAny,
		data:       nil,
		offset:     0,
		metadata:   nil,
		key:        nil,
		waitResult: false,
	}
	for _, o := range opts {
		o(opt)
	}
	if len(opt.topic) <= 0 {
		panic("topic empty")
	}
	if opt.p == nil {
		panic("kafka producer client should be initialised")
	}

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &opt.topic, Partition: opt.partition, Offset: opt.offset, Metadata: opt.metadata},
		Value:          opt.data,
		Key:            opt.key,
		Timestamp:      time.Now(),
		TimestampType:  kafka.TimestampCreateTime,
		Opaque:         nil, // not support
		Headers:        nil, // not support
	}
	var delivery chan kafka.Event
	if opt.waitResult {
		delivery = make(chan kafka.Event)
	}
	err := opt.p.Produce(message, delivery)
	if err != nil {
		panic(err)
	}

	if !opt.waitResult {
		return true
	}

	for e := range delivery {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				return false
			} else {
				return true
			}
		}
	}
	return false
}
