package pulsar_queue

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"task/config"
)

var (
	PulsarClient pulsar.Client
)

func NewPulsarClient(cfg config.PulsarConfig) (pulsar.Client, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.Url,
	})
	if err != nil {
		return nil, err
	}
	PulsarClient = client
	return PulsarClient, nil
}

func NewConsumer(consumerOptions pulsar.ConsumerOptions) (pulsar.Consumer, error) {
	consumer, err := PulsarClient.Subscribe(consumerOptions)
	if err != nil {
		return nil, err
	}
	return consumer, nil
}

func NewProducer(producerOptions pulsar.ProducerOptions) (pulsar.Producer, error) {
	producer, err := PulsarClient.CreateProducer(producerOptions)
	if err != nil {
		return nil, err
	}
	return producer, nil
}
