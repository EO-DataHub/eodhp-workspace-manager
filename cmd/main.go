package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/events"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/k8s"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/manager"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/rs/zerolog/log"
	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {

	// Initialize configuration
	appConfig := utils.LoadConfig()

	// Initialize logger
	utils.InitLogger(appConfig.LogLevel)
	log.Info().Msg("Workspace Manager starting...")

	// Initialize Pulsar client
	pulsarClient := events.NewPulsarClient(appConfig.Pulsar.URL)

	// Start Kubernetes manager to monitor 'Workspace' CR status changes
	k8sMgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes manager")
	}

	// Start the Kubernetes manager
	go func() {
		if err := k8sMgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Fatal().Err(err).Msg("Failed to start Kubernetes manager")
		}
	}()

	// Initialize WorkspaceController to manage Workspace CRUD operations
	k8sClient := k8s.NewClient(appConfig)
	workspaceController := manager.NewWorkspaceOperator(k8sClient, pulsarClient)

	// Start StatusPublisher to detect status changes and publish updates
	statusPublisher := events.NewStatusPublisher(k8sMgr.GetClient(), pulsarClient, appConfig.Pulsar.TopicProducer)
	go func() {
		if err := statusPublisher.Start(context.Background(), k8sMgr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start StatusWatcher")
		}
	}()

	// Start ConfigurationConsumer to process messages and apply CRUD operations
	configurationConsumer := events.NewConfigurationConsumer(pulsarClient, workspaceController, appConfig)
	go configurationConsumer.Start()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Shutting down Workspace Manager...")
	configurationConsumer.Stop()

}
