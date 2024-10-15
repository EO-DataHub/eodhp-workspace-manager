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

	// Setup logging
	initLogging(*logLevel)

	// Set up the Kubernetes client
	k8sClient, err := k8s.CreateK8sClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}

	// Setup the Pulsar client
	pulsarURL := "pulsar://localhost:6650"
	pulsarClient, err := events.NewPulsarClient(pulsarURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar client")
	}
	defer pulsarClient.Close()

	// Initialize the Manager with our clients
	mgr := manager.NewManager(k8sClient, pulsarClient)

	// Start Pulsar listener and pass the Manager to handle messages
	initPulsar(mgr)

	log.Info().Msg("Workspace Manager started")

}

func initPulsar(mgr *manager.Manager) {

	// TODO: Get Pulsar connection settings from environment variables in our ArgoCD deployment
	topic := "workspace"                 // Think of better name?
	subscription := "workspace-listener" // Think of better name?
	client := mgr.PulsarClient

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: subscription,
		Type:             pulsar.Shared,
		DLQ: &pulsar.DLQPolicy{
			MaxDeliveries:   1,                                 // TODO: optimize this in the next few sprints
			DeadLetterTopic: "persistent://public/default/dlq", // Dead letter topic - if message fails after MaxDeliveries attempts
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
	cancel()

}

func initLogging(logLevel string) {

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
