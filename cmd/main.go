package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-manager/events"
	"github.com/EO-DataHub/eodhp-workspace-manager/k8s"
	"github.com/EO-DataHub/eodhp-workspace-manager/manager"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	logLevel := flag.String("log", "info", "Set the logging level (debug, info, warn, error, fatal, panic)")

	flag.Parse()

	initLogging(*logLevel)

	// Set up the Kubernetes client using controller-runtime client
	k8sClient, err := k8s.CreateK8sClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Initialize Pulsar client using the NewPulsarClient function
	pulsarURL := "pulsar://localhost:6650"
	pulsarClient, err := events.NewPulsarClient(pulsarURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar client")
	}
	defer pulsarClient.Close()

	// Initialize the Manager with Kubernetes and Pulsar clients
	mgr := manager.NewManager(k8sClient, pulsarClient)

	// Start Pulsar listener and pass the Manager to handle messages

	initPulsar(mgr)

	log.Info().Msg("Application started")

}

func initPulsar(mgr *manager.Manager) {
	// Get Pulsar connection settings from environment variables

	topic := "workspace"
	subscription := "my-subscription"
	client := mgr.PulsarClient

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
		DLQ: &pulsar.DLQPolicy{
			MaxDeliveries:   1, // Maximum 1 attempts before sending to nacked topic
			DeadLetterTopic: "persistent://public/default/nacked",
		},
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Could not subscribe to topic")
	}
	defer consumer.Close()

	// Create a context for the message listener
	ctx, cancel := context.WithCancel(context.Background())

	// Start listening for messages using the manager
	go events.ListenForMessages(ctx, consumer, mgr)

	// Wait for termination signal (e.g., Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	// Trigger graceful shutdown
	log.Info().Msg("Received shutdown signal, stopping listener...")
	cancel() // Stop the listener

}

func initLogging(logLevel string) {
	// Set global time field format
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Set output to console
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

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
