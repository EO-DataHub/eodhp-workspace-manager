package events

import (
	"github.com/rs/zerolog/log"

	"github.com/apache/pulsar-client-go/pulsar"
)

func NewPulsarClient(pulsarURL string) pulsar.Client {
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: pulsarURL})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar client")
	}
	return client
}
