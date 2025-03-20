package pulsar_queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
)

const (
	pulsarURL        = "pulsar_queue://localhost:16650"
	testTopic        = "test-topic"
	testMessage      = "Hello Pulsar!"
	subscriptionName = "test-subscription"
)

func TestPulsarProducerConsumer(t *testing.T) {
	// 创建 Pulsar 客户端
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:               pulsarURL,
		OperationTimeout:  30 * time.Second,
		ConnectionTimeout: 30 * time.Second,
	})
	if err != nil {
		t.Fatalf("Could not create pulsar_queue client: %v", err)
	}
	defer client.Close()

	// 创建消费者
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            testTopic,
		SubscriptionName: subscriptionName,
		Type:             pulsar.Shared,
	})
	if err != nil {
		t.Fatalf("Could not create consumer: %v", err)
	}
	defer consumer.Close()

	// 创建生产者
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: testTopic,
	})
	if err != nil {
		t.Fatalf("Could not create producer: %v", err)
	}
	defer producer.Close()

	// 发送消息
	msgID, err := producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: []byte(testMessage),
	})
	fmt.Printf("消息id: %s\n", msgID)
	assert.NoError(t, err)
	assert.NotNil(t, msgID)

	// 接收消息
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	msg, err := consumer.Receive(ctx)
	if err != nil {
		t.Fatalf("Could not receive message: %v", err)
	}

	// 验证消息内容
	assert.Equal(t, testMessage, string(msg.Payload()))

	// 确认消息
	consumer.Ack(msg)
}
