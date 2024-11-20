package k8s

import (
	"context"
	"fmt"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MapObjectStoresToS3Buckets maps ObjectStores to S3Buckets
func (c *K8sClient) MapObjectStoresToS3Buckets(workspaceName string, objectStores []models.ObjectStore) []workspacev1alpha1.S3Bucket {
	var buckets []workspacev1alpha1.S3Bucket
	for _, obj := range objectStores {
		buckets = append(buckets, workspacev1alpha1.S3Bucket{
			Name:            obj.Name,
			Path:            obj.Path,
			EnvVar:          obj.EnvVar,
			AccessPointName: fmt.Sprintf("%s-%s-s3", c.Config.AWS.Cluster, workspaceName),
		})
	}

	return buckets
}

// MapBlockStoresToEFSAccessPoints maps BlockStores to EFSAccessPoints
func (c *K8sClient) MapBlockStoresToEFSAccessPoints(workspaceName string, blockStores []models.BlockStore) []workspacev1alpha1.EFSAccess {
	var accessPoints []workspacev1alpha1.EFSAccess
	for _, block := range blockStores {
		accessPoints = append(accessPoints, workspacev1alpha1.EFSAccess{
			Name:          block.Name,
			FSID:          c.Config.AWS.FSID,
			RootDirectory: "/workspaces/" + workspaceName,
			User: workspacev1alpha1.User{
				UID: 1000, // Default UID
				GID: 1000, // Default GID
			},
			Permissions: "0755", // Default permissions
		})
	}
	return accessPoints
}

// GenerateStorageConfig generates a StorageSpec for a Workspace based on the workspace name
func (c *K8sClient) GenerateStorageConfig(workspaceName string) workspacev1alpha1.StorageSpec {
	pvName := fmt.Sprintf("pv-%s-workspace", workspaceName)

	return workspacev1alpha1.StorageSpec{
		PersistentVolumes: []workspacev1alpha1.PVSpec{
			{
				Name:         pvName,
				StorageClass: c.Config.Storage.StorageClass,
				Size:         c.Config.Storage.Size,
				VolumeSource: &workspacev1alpha1.VolumeSource{
					Driver:          c.Config.Storage.Driver,
					AccessPointName: fmt.Sprintf("%s-%s-pv", c.Config.AWS.Cluster, workspaceName),
				},
			},
		},
		PersistentVolumeClaims: []workspacev1alpha1.PVCSpec{
			{
				PVSpec: workspacev1alpha1.PVSpec{
					Name:         c.Config.Storage.PVCName,
					StorageClass: c.Config.Storage.StorageClass,
					Size:         c.Config.Storage.Size,
				},
				PVName: pvName,
			},
		},
	}
}

// buildWorkspace creates a Workspace object based on the provided WorkspaceSettings
func (c *K8sClient) buildWorkspace(req models.WorkspaceSettings) *workspacev1alpha1.Workspace {
	var s3Buckets []workspacev1alpha1.S3Bucket
	var efsAccessPoints []workspacev1alpha1.EFSAccess

	if req.Stores != nil {
		for _, store := range *req.Stores {
			// Map ObjectStores to S3Buckets
			s3Buckets = append(s3Buckets, c.MapObjectStoresToS3Buckets(req.Name, store.Object)...)
			// Map BlockStores to EFSAccessPoints
			efsAccessPoints = append(efsAccessPoints, c.MapBlockStoresToEFSAccessPoints(req.Name, store.Block)...)
		}
	}

	// Generate storage configuration based on workspace name
	storageConfig := c.GenerateStorageConfig(req.Name)

	// Create the Workspace object
	return &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: "workspaces",
		},
		Spec: workspacev1alpha1.WorkspaceSpec{
			Namespace: "ws-" + req.Name,
			Account:   req.Account.String(),
			Authorization: workspacev1alpha1.AuthorizationSpec{
				MemberGroup: req.MemberGroup,
			},
			AWS: workspacev1alpha1.AWSSpec{
				RoleName: fmt.Sprintf("%s-%s", c.Config.AWS.Cluster, req.Name),
				EFS: workspacev1alpha1.EFSSpec{
					AccessPoints: efsAccessPoints,
				},
				S3: workspacev1alpha1.S3Spec{
					Buckets: s3Buckets,
				},
			},
			ServiceAccount: workspacev1alpha1.ServiceAccountSpec{
				Name: "default",
			},
			Storage: storageConfig,
		},
	}
}

// CreateWorkspace creates a new Workspace in the cluster
func (c *K8sClient) CreateWorkspace(ctx context.Context, req models.WorkspaceSettings) error {
	workspace := c.buildWorkspace(req)

	err := c.client.Create(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to create workspace %s: %w", req.Name, err)
	}

	log.Info().Str("name", req.Name).Str("namespace", req.MemberGroup).Msg("Workspace successfully created")
	return nil
}

// UpdateWorkspace updates an existing Workspace in the cluster
func (c *K8sClient) UpdateWorkspace(ctx context.Context, req models.WorkspaceSettings) error {

	// Retrieve the existing Workspace from the cluster
	existingWorkspace := &workspacev1alpha1.Workspace{}
	err := c.client.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: "workspaces"}, existingWorkspace)
	if err != nil {
		return fmt.Errorf("failed to fetch existing workspace %s: %w", req.Name, err)
	}

	// Build the updated Workspace
	updatedWorkspace := c.buildWorkspace(req)

	// Set the ResourceVersion to ensure the update is successful
	updatedWorkspace.ObjectMeta.ResourceVersion = existingWorkspace.ObjectMeta.ResourceVersion

	// Perform the update operation
	err = c.client.Update(ctx, updatedWorkspace)
	if err != nil {
		return fmt.Errorf("failed to update workspace %s: %w", req.Name, err)
	}

	log.Info().Str("name", req.Name).Str("namespace", req.MemberGroup).Msg("Workspace successfully updated")
	return nil
}

// DeleteWorkspace deletes an existing Workspace in the cluster
func (c *K8sClient) DeleteWorkspace(ctx context.Context, payload models.WorkspaceSettings) error {

	// Define the workspace to delete
	workspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      payload.Name,
			Namespace: "workspaces",
		},
	}

	// Attempt to delete the workspace
	err := c.client.Delete(ctx, workspace)
	if err != nil {
		return fmt.Errorf("failed to delete workspace %s: %w", payload.Name, err)
	}

	log.Info().Str("name", payload.Name).Msg("Workspace successfully deleted")
	return nil

}
