package k8s

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type WorkspaceWatcher struct {
	k8sClient    client.Client
	pulsarClient pulsar.Client
	topic        string
}

func NewWorkspaceWatcher(k8sClient client.Client, pulsarClient pulsar.Client, topic string) *WorkspaceWatcher {
	return &WorkspaceWatcher{
		k8sClient:    k8sClient,
		pulsarClient: pulsarClient,
		topic:        topic,
	}
}

func (w *WorkspaceWatcher) Start(ctx context.Context, mgr manager.Manager) error {
	informer, err := mgr.GetCache().GetInformer(ctx, &workspacev1alpha1.Workspace{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create informer for Workspace CR")
		return err
	}

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: w.handleUpdate,
	})

	log.Info().Msg("Started WorkspaceWatcher")
	return nil
}

func (w *WorkspaceWatcher) handleUpdate(oldObj, newObj interface{}) {
	oldWorkspace, ok := oldObj.(*workspacev1alpha1.Workspace)
	if !ok {
		log.Error().Msg("Failed to cast old object to Workspace")
		return
	}

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

	// Log status change detection
	log.Info().Msg("STATUS HAS CHANGED!")

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

	// Send the status update to Pulsar
	if err := w.sendStatusUpdate(statusBytes); err != nil {
		log.Error().Err(err).Msg("Failed to send status update to Pulsar")
	}
}

func (w *WorkspaceWatcher) sendStatusUpdate(statusBytes []byte) error {
	producer, err := w.pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: w.topic,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Pulsar producer")
		return err
	}
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: statusBytes,
	})
	return err
}
