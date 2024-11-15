package events

import (
	"context"
	"encoding/json"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/manager"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
)

type Listener struct {
	Consumer pulsar.Consumer
	Operator *manager.ResourceOperator
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewListener creates a new Pulsar listener
func NewListener(client pulsar.Client, topic, subscription string, operator *manager.ResourceOperator) *Listener {
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar consumer")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &Listener{
		Consumer: consumer,
		Operator: operator,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins listening for messages
func (l *Listener) Start() {
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

// Stop gracefully shuts down the listener
func (l *Listener) Stop() {
	l.cancel()
	l.Consumer.Close()
}

func (l *Listener) handleMessage(msg pulsar.Message) {
	var payload models.WorkspacePayload
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
