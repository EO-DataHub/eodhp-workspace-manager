package cmd

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/k8s"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	configFile string
	rootCmd    = &cobra.Command{
		Use:   "workspace-manager",
		Short: "Workspace Manager CLI",
		Long:  "A CLI to manage workspaces with Kubernetes and Pulsar integration.",
		Run:   runWorkspaceManager,
	}
)

// init initializes the root command
func init() {
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
}

func Execute() {
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}

// runWorkspaceManager is the main entry point for the Workspace Manager
func runWorkspaceManager(cmd *cobra.Command, args []string) {
	// Load configuration
	appConfig := utils.LoadConfig(configFile)

	// Initialize logger
	utils.InitLogger(appConfig.LogLevel)
	log.Info().Msg("Workspace Manager starting...")

	// Initialize Pulsar client
	pulsarClient, err := pulsar.NewClient(pulsar.ClientOptions{URL: appConfig.Pulsar.URL, MaxConnectionsPerBroker: 1})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar Client")
	}
	defer pulsarClient.Close()

	// Producer for workspace-status topic
	statusProducer, err := pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: appConfig.Pulsar.TopicProducer,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar producer for workspace-status")
	}
	defer statusProducer.Close()

	// Consumer for workspace-settings topic
	settingsConsumer, err := pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:            appConfig.Pulsar.TopicConsumer,
		SubscriptionName: appConfig.Pulsar.Subscription,
		Type:             pulsar.Shared,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pulsar consumer for workspace-settings")
	}
	defer settingsConsumer.Close()

	// Initialize Kubernetes manager
	k8sMgr, err := k8s.InitializeManager()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Kubernetes manager")
	}

	// Start the Kubernetes manager
	go func() {
		if err := k8sMgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Fatal().Err(err).Msg("Failed to start Kubernetes manager")
		}
	}()

	// Listen for updates to workspace CR status and send updates to workspace-status topic
	chanWorkspaceStatus := make(chan models.WorkspaceStatus, 100)
	if err := k8s.ListenForWorkspaceStatusUpdates(context.Background(), k8sMgr, chanWorkspaceStatus); err != nil {
		log.Fatal().Err(err).Msg("Failed to start informer")
	}

	//Start the consumer loop to process workspace-settings messages
	go func() {
		for {
			msg, err := settingsConsumer.Receive(context.Background())
			if err != nil {
				log.Error().Err(err).Msg("Error receiving message from Pulsar")
				panic(err)
			}

			// Parse the message into WorkspaceSettings
			var payload models.WorkspaceSettings
			if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal workspace-settings message")
				settingsConsumer.Nack(msg)
				continue
			}

			// Process the workspace settings message
			if err := k8s.ProcessWorkspace(context.Background(), k8sMgr.GetClient(), appConfig, payload); err != nil {
				log.Error().Err(err).Msg("Failed to process workspace settings message")
				settingsConsumer.Nack(msg)
			} else {
				settingsConsumer.Ack(msg)
				log.Info().Msg("Message successfully processed and acknowledged")
			}

		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info().Msg("Shutting down Workspace Manager...")
}
