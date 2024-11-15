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

	// Initialize configuration and logging
	config := utils.LoadConfig()
	utils.InitLogger(config.LogLevel)
	log.Info().Msg("Workspace Manager starting...")

	// Initialize Pulsar client
	pulsarClient := events.NewPulsarClient(config.Pulsar.URL)

	// Start Kubernetes manager to watch for changes to Workspace resources
	k8sMgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes manager")
	}

	log.Info().Msg(k8sMgr.GetConfig().APIPath)

	// Initialize ResourceOperator to allow making changes to Workspace Resources
	k8sClient := k8s.NewClient()
	resourceOperator := manager.NewResourceOperator(k8sClient, pulsarClient)

	// Start Pulsar listener to listen for messages in workspace-settings topic
	listener := events.NewListener(pulsarClient, config.Pulsar.TopicProducer, config.Pulsar.Subscription, resourceOperator)
	go listener.Start()

	// Start StatusWatcher
	statusWatcher := k8s.NewWorkspaceWatcher(k8sMgr.GetClient(), pulsarClient, config.Pulsar.TopicConsumer)
	go func() {
		if err := statusWatcher.Start(context.Background(), k8sMgr); err != nil {
			log.Fatal().Err(err).Msg("Failed to start StatusWatcher")
		}
	}()

	// Start the Kubernetes manager
	go func() {
		if err := k8sMgr.Start(ctrl.SetupSignalHandler()); err != nil {
			log.Fatal().Err(err).Msg("Failed to start Kubernetes manager")
		}
	}()

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info().Msg("Shutting down Workspace Manager...")
	listener.Stop()

}
