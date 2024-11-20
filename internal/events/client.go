package events

import (
	"github.com/rs/zerolog/log"

	"github.com/apache/pulsar-client-go/pulsar"
)

// Client defines the interface for a Pulsar client
type PulsarClient interface {
	CreateProducer(options pulsar.ProducerOptions) (pulsar.Producer, error)
	Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error)
	Close()
}

// NewPulsarClient creates a new Pulsar client
func NewPulsarClient(pulsarURL string) PulsarClient {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: pulsarURL})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar client")
	}
	return client
}

// MockPulsarClient is a mock implementation of PulsarClient
type MockPulsarClient struct {
	SubscribeFunc      func(options pulsar.ConsumerOptions) (pulsar.Consumer, error)
	CreateProducerFunc func(options pulsar.ProducerOptions) (pulsar.Producer, error)
	CloseFunc          func()
}

// Subscribe subscribes to a Pulsar topic
func (m *MockPulsarClient) Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error) {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(options)
	}
	return nil, nil
}

// CreateProducer creates a Pulsar producer
func (m *MockPulsarClient) CreateProducer(options pulsar.ProducerOptions) (pulsar.Producer, error) {
	if m.CreateProducerFunc != nil {
		return m.CreateProducerFunc(options)
	}
	return nil, nil
}

// Close closes the Pulsar client
func (m *MockPulsarClient) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}
