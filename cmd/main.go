package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	// workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"

	"github.com/EO-DataHub/eodhp-workspace-manager/manager"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func listenForMessages(ctx context.Context, consumer pulsar.Consumer) {
	log.Info().Msg("Message listener started...")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Shutting down message listener...")
			return
		default:
			msg, err := consumer.Receive(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Failed to receive message")
				time.Sleep(1 * time.Second) // Simple backoff in case of failure
				continue
			}

			log.Info().Msg("Received message: " + string(msg.Payload()))

			// Pass the message to the manager package for handling
			err = manager.HandleMessage(msg)
			if err != nil {
				log.Printf("Failed to handle message: %v", err)
				consumer.Nack(msg) // Negative Acknowledge the message
				continue
			}

			consumer.Ack(msg) // Acknowledge the message if handled successfully
		}
	}
}

func main() {

	logLevel := flag.String("log-level", "info", "Set the logging level (debug, info, warn, error, fatal, panic)")

	// Parse the command-line flags
	flag.Parse()

	initLogging(*logLevel)
	initPulsar()

	log.Info().Msg("Application started")

}

func initPulsar() {
	// Get Pulsar connection settings from environment variables
	pulsarURL := "pulsar://localhost:6650"
	topic := "workspace"
	subscription := "my-subscription"

	// Initialize Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not create Pulsar client")
	}
	defer client.Close()

	// Set up a consumer
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared, // Shared subscription model
		DLQ: &pulsar.DLQPolicy{
			MaxDeliveries:   3, // Maximum 3 attempts before sending to nacked-messages topic
			DeadLetterTopic: "persistent://public/default/nacked-messages",
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not subscribe to topic")
	}
	defer consumer.Close()

	// Create a context to manage the lifecycle of the message listener
	ctx, cancel := context.WithCancel(context.Background())

	// Start the message listener in a separate goroutine
	go listenForMessages(ctx, consumer)

	// Wait for a signal (e.g., Ctrl+C) to gracefully shut down the application
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Trigger graceful shutdown
	log.Info().Msg("Received shutdown signal, stopping listener...")
	cancel() // Cancel the context to stop the listener
}
func initLogging(logLevel string) {
	// Set global time field format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set output to console for human-readable logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Map the log level string to Zerolog log levels
	switch logLevel {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		log.Warn().Str("level", logLevel).Msg("Invalid log level, defaulting to info")
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	log.Info().Str("log_level", logLevel).Msg("Log level set")
}
