package manager

import (
	"context"
	"encoding/json"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/k8s"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
)

// ResourceOperator handles applying operations on Kubernetes resources
type ResourceOperator struct {
	K8sClient    *k8s.Client
	PulsarClient pulsar.Client
}

// NewResourceOperator initializes a new ResourceOperator
func NewResourceOperator(k8sClient *k8s.Client, pulsarClient pulsar.Client) *ResourceOperator {
	return &ResourceOperator{
		K8sClient:    k8sClient,
		PulsarClient: pulsarClient,
	}
}

// ProcessMessage handles an incoming Pulsar message and applies operations to Kubernetes
func (ro *ResourceOperator) ProcessMessage(ctx context.Context, payload models.WorkspacePayload) error {
	switch payload.Action {
	case "create":
		log.Info().Str("workspaceName", payload.Name).Msg("Creating workspace")
		//return ro.K8sClient.CreateWorkspace(ctx, payload)
	case "update":
		log.Info().Str("workspaceName", payload.Name).Msg("Updating workspace")
		//return ro.K8sClient.UpdateWorkspace(ctx, payload)
	case "delete":
		log.Info().Str("workspaceName", payload.Name).Msg("Deleting workspace")
		//return ro.K8sClient.DeleteWorkspace(ctx, payload)
	case "patch":
		log.Info().Str("workspaceName", payload.Name).Msg("Patching workspace")
		//return ro.K8sClient.PatchWorkspace(ctx, payload)
	default:
		log.Error().Str("action", payload.Action).Msg("Unknown action in Pulsar message")
		return nil
	}

	return nil
}

// NotifyStatusChange publishes updates to the Workspace status to Pulsar
func (ro *ResourceOperator) NotifyStatusChange(workspace *workspacev1alpha1.Workspace) error {
	statusUpdate := map[string]interface{}{
		"workspaceName": workspace.Name,
		"namespace":     workspace.Namespace,
		"status":        workspace.Status,
	}

	payload, err := json.Marshal(statusUpdate)
	if err != nil {
		log.Error().Err(err).Msg("Failed to serialize status update")
		return err
	}

	return ro.sendToPulsar(payload)
}

func (ro *ResourceOperator) sendToPulsar(payload []byte) error {
	producer, err := ro.PulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: "workspace-status", // Replace with your actual topic
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Pulsar producer")
		return err
	}
	defer producer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = producer.Send(ctx, &pulsar.ProducerMessage{
		Payload: payload,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to send status update to Pulsar")
		return err
	}

	log.Info().Msg("Successfully sent status update to Pulsar")
	return nil
}
