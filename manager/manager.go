package manager

import (
	"context"
	"encoding/json"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	pulsar "github.com/apache/pulsar-client-go/pulsar"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager holds shared dependencies (K8s and Pulsar clients)
type Manager struct {
	K8sClient    client.Client
	PulsarClient pulsar.Client
}

// NewManager creates a new Manager with K8s and Pulsar clients
func NewManager(k8sClient client.Client, pulsarClient pulsar.Client) *Manager {
	return &Manager{
		K8sClient:    k8sClient,
		PulsarClient: pulsarClient,
	}
}

// HandleMessage processes an incoming Pulsar message and applies Kubernetes CRUD operations
func (m *Manager) HandleMessage(msg pulsar.Message) error {
	var workspaceMsg models.WorkspaceMessage
	err := json.Unmarshal(msg.Payload(), &workspaceMsg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse message payload")
		return err
	}

	switch workspaceMsg.Action {
	case "create":
		return m.createWorkspace(&workspaceMsg.Spec)
	// case "update":
	// 	return m.updateWorkspace(&workspaceMsg.Spec)
	// case "delete":
	// 	return m.deleteWorkspace(workspaceMsg.Spec.Namespace, workspaceMsg.Spec.Namespace)
	default:
		log.Error().Str("action", workspaceMsg.Action).Msg("Unknown action")
		return err
	}
}
func mapEFSAccessPoints(accessPoints []models.AWSEFSAccessPoint) []workspacev1alpha1.EFSAccess {
	var efsAccessPoints []workspacev1alpha1.EFSAccess
	for _, ap := range accessPoints {
		efsAccessPoints = append(efsAccessPoints, workspacev1alpha1.EFSAccess{
			Name:          ap.Name,
			FSID:          ap.FSID,
			RootDirectory: ap.RootDir,
			User: workspacev1alpha1.User{
				UID: int64(ap.UID),
				GID: int64(ap.GID),
			},
			Permissions: ap.Permissions,
		})
	}
	return efsAccessPoints
}

func (m *Manager) createWorkspace(req *models.WorkspaceRequest) error {

	workspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: *models.WorkspaceRequestToSpec(req),
	}

	workspaceJSON, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal workspace object to JSON")
	} else {
		// Log the workspace object as JSON
		log.Info().Msgf("Workspace object: %s", string(workspaceJSON))
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the workspace already exists
	existingWorkspace := &workspacev1alpha1.Workspace{}
	err = m.K8sClient.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, existingWorkspace)
	if err == nil {
		log.Info().Str("name", req.Name).Msg("Workspace already exists")
		return nil
	}
	log.Info().Str("name", req.Name).Msg("Workspace DOES NOT exists")
	// workspace := &workspacev1alpha1.Workspace{}
	// err := r.Client.Get(ctx, client.ObjectKey{Name: workspace.Name, Namespace: workspace.Namespace}, existingWorkspace)
	// if err == nil {
	// 	log.Info().Str("name", workspace.Name).Msg("Workspace already exists")
	// 	return nil // Return if workspace exists, no need to create
	// }

	// workspace := &workspacev1alpha1.Workspace{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      spec.Namespace,
	// 		Namespace: spec.Namespace,
	// 	},
	// 	Spec: *spec,
	// }

	// _, err := m.K8sClient.Resource(models.WorkspaceGVR).Namespace(spec.Namespace).Create(context.TODO(), workspace, metav1.CreateOptions{})
	// if err != nil {
	// 	log.Error().Err(err).Msg("Failed to create Workspace")
	// 	return err
	// }

	log.Info().Str("workspace", req.Name).Msg("Successfully created Workspace")
	return nil
}

// func (m *Manager) updateWorkspace(spec *workspacev1alpha1.WorkspaceSpec) error {
// 	workspace := &workspacev1alpha1.Workspace{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      spec.Namespace,
// 			Namespace: spec.Namespace,
// 		},
// 		Spec: *spec,
// 	}

// 	_, err := m.K8sClient.Resource(models.WorkspaceGVR).Namespace(spec.Namespace).Update(context.TODO(), workspace, metav1.UpdateOptions{})
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to update Workspace")
// 		return err
// 	}

// 	log.Info().Str("workspace", spec.Namespace).Msg("Successfully updated Workspace")
// 	return nil
// }

// func (m *Manager) deleteWorkspace(name, namespace string) error {
// 	err := m.K8sClient.Resource(models.WorkspaceGVR).Namespace(namespace).Delete(context.TODO(), name, metav1.DeleteOptions{})
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to delete Workspace")
// 		return err
// 	}

// 	log.Info().Str("workspace", name).Msg("Successfully deleted Workspace")
// 	return nil
// }
