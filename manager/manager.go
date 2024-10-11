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
	var WorkspacePayload models.WorkspacePayload
	err := json.Unmarshal(msg.Payload(), &WorkspacePayload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse message payload")
		return err
	}

	switch WorkspacePayload.Action {
	case "create":
		return m.createWorkspace(&WorkspacePayload)
	// case "update":
	// 	return m.updateWorkspace(&workspaceMsg.Spec)
	case "delete":
		return m.deleteWorkspace(&WorkspacePayload)
	default:
		log.Error().Str("action", WorkspacePayload.Action).Msg("Unknown action")
		return err
	}
}

func (m *Manager) createWorkspace(req *models.WorkspacePayload) error {

	// Create a new Workspace object to be created in the cluster
	workspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
		Spec: workspacev1alpha1.WorkspaceSpec{
			Namespace: req.Namespace,
			AWS: workspacev1alpha1.AWSSpec{
				RoleName: *req.AWSRoleName, // AWS RoleName, assuming it's not nil
				EFS: workspacev1alpha1.EFSSpec{
					AccessPoints: models.MapEFSAccessPoints(req.EFSAccessPoint),
				},
				S3: workspacev1alpha1.S3Spec{
					Buckets: models.MapS3Buckets(req.S3Buckets),
				},
			},
			ServiceAccount: workspacev1alpha1.ServiceAccountSpec{
				Name: *req.ServiceAccountName, // Service Account, assuming it's not nil
			},
			Storage: workspacev1alpha1.StorageSpec{
				PersistentVolumes:      models.MapPersistentVolumes(req.PersistentVolumes),
				PersistentVolumeClaims: models.MapPersistentVolumeClaims(req.PersistentVolumeClaims),
			},
		},
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

	log.Info().Str("name", req.Name).Msg("Workspace does not exists. Creating it now.")
	err = m.K8sClient.Create(ctx, workspace)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create Workspace")
		return err
	}

	log.Info().Str("name", req.Name).Msg("Successfully created Workspace")
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

func (m *Manager) deleteWorkspace(req *models.WorkspacePayload) error {

	workspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
		},
	}

	// Set a timeout for the operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if the workspace exists
	existingWorkspace := &workspacev1alpha1.Workspace{}
	err := m.K8sClient.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, existingWorkspace)
	if err != nil {
		// If the workspace does not exist, log it and return
		log.Info().Str("name", req.Name).Str("namespace", req.Namespace).Msg("Workspace does not exist. Nothing to delete.")
		return nil
	}
	log.Info().Str("name", req.Name).Str("namespace", req.Namespace).Msg("Deleting workspace.")

	err = m.K8sClient.Delete(ctx, workspace)
	if err != nil {
		log.Error().Err(err).Str("name", req.Name).Str("namespace", req.Namespace).Msg("Failed to delete Workspace")
		return err
	}

	log.Info().Str("workspace", req.Name).Msg("Successfully deleted Workspace")
	return nil
}
