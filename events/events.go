package events

import (
	"context"

	"github.com/EO-DataHub/eodhp-workspace-manager/manager" // Import the manager package to access HandleMessage
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
)

func NewPulsarClient(pulsarURL string) (pulsar.Client, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Pulsar client")
		return nil, err
	}
	return client, nil
}

// Listens for messages from the Pulsar consumer and passes them to the Manager
func ListenForMessages(ctx context.Context, consumer pulsar.Consumer, mgr *manager.Manager) {
	log.Info().Msg("Listening for messages...")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Stopping message listener...")
			return
		default:
			// Receive and handle Pulsar message
			msg, err := consumer.Receive(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to receive message")
				continue
			}
			err = mgr.HandleMessage(msg)
			if err != nil {
				consumer.Nack(msg) 
			} else {
				consumer.Ack(msg)
			}
		}
	}
}
