package events

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// StatusPublisher publishes Workspace status updates to a Pulsar topic
type StatusPublisher struct {
	k8sClient    client.Client
	pulsarClient PulsarClient
	topic        string
}

// NewStatusPublisher initializes a new StatusPublisher
func NewStatusPublisher(k8sClient client.Client, pulsarClient PulsarClient, topic string) *StatusPublisher {
	return &StatusPublisher{
		k8sClient:    k8sClient,
		pulsarClient: pulsarClient,
		topic:        topic,
	}
}

// Start starts the StatusPublisher
func (w *StatusPublisher) Start(ctx context.Context, mgr manager.Manager) error {
	informer, err := mgr.GetCache().GetInformer(ctx, &workspacev1alpha1.Workspace{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create informer for Workspace CR")
		return err
	}

	// Add event handler to informer
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.handleUpdate,
	})

	log.Info().Msg("Listening for Workspace Status Updates...")
	return nil
}

// handleUpdate is called when a Workspace CR status is updated
func (w *StatusPublisher) handleUpdate(oldObj, newObj interface{}) {
	oldWorkspace, ok := oldObj.(*workspacev1alpha1.Workspace)
	if !ok {
		log.Error().Msg("Failed to cast old object to Workspace")
		return
	}

	// Cast new object to Workspace
	newWorkspace, ok := newObj.(*workspacev1alpha1.Workspace)
	if !ok {
		log.Error().Msg("Failed to cast new object to Workspace")
		return
	}

	// Check if the status has actually changed
	if reflect.DeepEqual(oldWorkspace.Status, newWorkspace.Status) {
		// If status hasn't changed, ignore the event
		log.Debug().Msg("No changes in Workspace status; skipping")
		return
	}

	// Prepare the status update message
	statusUpdate := map[string]interface{}{
		"workspaceName": newWorkspace.Name,
		"namespace":     newWorkspace.Namespace,
		"status":        newWorkspace.Status,
	}
	statusBytes, err := json.Marshal(statusUpdate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize status update")
		return
	}

	log.Info().Msg("Status Update..")

	// Send the status update to Pulsar
	if err := w.sendStatusUpdate(statusBytes); err != nil {
		log.Error().Err(err).Msg("Failed to send status update to Pulsar")
	}
}

// sendStatusUpdate sends a status update to a Pulsar topic
func (w *StatusPublisher) sendStatusUpdate(statusBytes []byte) error {
	producer, err := w.pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: w.topic,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Pulsar producer")
		return err
	}
	defer producer.Close()

	// Send the message
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: statusBytes,
	})
	return err
}
