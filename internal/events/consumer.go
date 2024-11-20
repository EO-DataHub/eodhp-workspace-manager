package events

import (
	"context"
	"encoding/json"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
)

// ConsumerInterface defines the methods required for a Pulsar consumer
type ConsumerInterface interface {
	Receive(ctx context.Context) (pulsar.Message, error)
	Ack(msg pulsar.Message) error
	Nack(msg pulsar.Message)
	Close()
}

// WorkspaceOperatorInterface defines the methods required for processing messages
type WorkspaceOperatorInterface interface {
	ProcessMessage(ctx context.Context, payload models.WorkspaceSettings) error
}

// ConfigurationConsumer listens for messages on a Pulsar topic
type ConfigurationConsumer struct {
	Consumer ConsumerInterface
	Operator WorkspaceOperatorInterface
	Config   *utils.Config
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConfigurationConsumer creates a new Pulsar ConfigurationConsumer
func NewConfigurationConsumer(client PulsarClient, operator WorkspaceOperatorInterface, config *utils.Config) *ConfigurationConsumer {
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            config.Pulsar.TopicConsumer,
		SubscriptionName: config.Pulsar.Subscription,
		Type:             pulsar.Shared,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar consumer")
	}

	// Create a new context for the consumer
	ctx, cancel := context.WithCancel(context.Background())
	return &ConfigurationConsumer{
		Consumer: consumer,
		Operator: operator,
		Config:   config,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins listening for messages
func (l *ConfigurationConsumer) Start() {
	log.Info().Msg("Starting Pulsar listener...")
	go func() {
		for {
			select {
			case <-l.ctx.Done():
				log.Info().Msg("Stopping Pulsar listener...")
				return
			default:
				msg, err := l.Consumer.Receive(l.ctx)
				if err != nil {
					log.Error().Err(err).Msg("Failed to receive Pulsar message")
					continue
				}

				// Process the message
				l.handleMessage(msg)
			}
		}
	}()
}

// Stop gracefully shuts down the ConfigurationConsumer
func (l *ConfigurationConsumer) Stop() {
	l.cancel()
	l.Consumer.Close()
}

// handleMessage processes a Pulsar message
func (l *ConfigurationConsumer) handleMessage(msg pulsar.Message) {

	// Unmarshal the message
	var payload models.WorkspaceSettings
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal Pulsar message payload")
		l.Consumer.Nack(msg)
		return
	}

	// Process the message using the ResourceOperator
	err = l.Operator.ProcessMessage(l.ctx, payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to process Pulsar message")
		l.Consumer.Nack(msg)
	} else {
		l.Consumer.Ack(msg)
	}
}
