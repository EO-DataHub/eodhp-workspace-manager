package manager

import (
	"context"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/k8s"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
)

// WorkspaceOperator manages operations on Kubernetes resources
type WorkspaceOperator struct {
	K8sClient    *k8s.Client
	PulsarClient pulsar.Client
}

// NewWorkspaceOperator creates a new WorkspaceOperator
func NewWorkspaceOperator(k8sClient *k8s.Client, pulsarClient pulsar.Client) *WorkspaceOperator {
	return &WorkspaceOperator{
		K8sClient:    k8sClient,
		PulsarClient: pulsarClient,
	}
}

// ProcessMessage processes a message from a Pulsar topic
func (ro *WorkspaceOperator) ProcessMessage(ctx context.Context, payload models.WorkspaceSettings) error {

	switch payload.Status {
	case "creating":
		return ro.K8sClient.CreateWorkspace(ctx, payload)
	case "updating":
		return ro.K8sClient.UpdateWorkspace(ctx, payload)
	case "delete":
		log.Info().Str("workspaceName", payload.Name).Msg("Deleting workspace")
		//return ro.K8sClient.DeleteWorkspace(ctx, payload)
	default:
		log.Error().Str("status", payload.Status).Msg("Unknown action in Pulsar message")
		return nil
	}

	return nil
}
