package k8s

import (
	"context"
	"testing"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// setupFakeClient creates a fake Kubernetes client and a mock application configuration
func setupFakeClient() (client.Client, *utils.Config) {
	// Add the Workspace CRD to the scheme
	s := scheme.Scheme
	_ = workspacev1alpha1.AddToScheme(s)

	// Create a fake Kubernetes client
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Mock application configuration
	mockConfig := &utils.Config{
		AWS: utils.AWSConfig{
			Cluster: "test-cluster",
			FSID:    "fs-test",
			Bucket:  "test-bucket",
		},
		Storage: utils.StorageConfig{
			Size:         "10Gi",
			StorageClass: "test-storage",
			PVCName:      "workspace-pvc",
			Driver:       "efs.csi.aws.com",
		},
	}

	return fakeClient, mockConfig
}

func TestMapObjectStoresToS3Buckets(t *testing.T) {

	// Setup test client and fake client builder
	_, c := setupFakeClient()
	objectStores := []models.ObjectStore{
		{Name: "obj1"},
		{Name: "obj2"},
	}

	// Call MapObjectStoresToS3Buckets
	result := MapObjectStoresToS3Buckets("test-workspace", c, objectStores)

	assert.Len(t, result, 2)
	assert.Equal(t, c.AWS.Bucket, result[0].Name)
	assert.Equal(t, "obj1/", result[0].Path)
	assert.Equal(t, "test-cluster-test-workspace-s3", result[0].AccessPointName)
}

func TestMapBlockStoresToEFSAccessPoints(t *testing.T) {

	// Setup test client and fake client builder
	_, c := setupFakeClient()

	blockStores := []models.BlockStore{
		{Name: "block1"},
		{Name: "block2"},
	}

	// Call MapBlockStoresToEFSAccessPoints
	result := MapBlockStoresToEFSAccessPoints("test-workspace", c, blockStores)

	assert.Len(t, result, 2)
	assert.Equal(t, "block1", result[0].Name)
	assert.Equal(t, "fs-test", result[0].FSID)
	assert.Equal(t, "/workspaces/test-workspace", result[0].RootDirectory)
	assert.Equal(t, "755", result[0].Permissions)
	assert.Equal(t, int64(1000), result[0].User.UID)
}

func TestGenerateStorageConfig(t *testing.T) {

	// Setup test client and fake client builder
	_, c := setupFakeClient()

	// Call GenerateStorageConfig
	result := GenerateStorageConfig("test-workspace", c)

	assert.Len(t, result.PersistentVolumes, 1)
	assert.Equal(t, "pv-test-workspace-workspace", result.PersistentVolumes[0].Name)
	assert.Equal(t, "test-storage", result.PersistentVolumes[0].StorageClass)
	assert.Equal(t, "10Gi", result.PersistentVolumes[0].Size)
	assert.Equal(t, "efs.csi.aws.com", result.PersistentVolumes[0].VolumeSource.Driver)
}

func TestCreateWorkspace(t *testing.T) {

	// Setup test client and fake client builder
	client, c := setupFakeClient()

	workspaceSettings := models.WorkspaceSettings{
		Name:    "test-workspace",
		Account: uuid.New(),
	}

	// Call CreateWorkspace
	err := CreateWorkspace(context.Background(), client, workspaceSettings, c)
	assert.NoError(t, err)

	// Verify the workspace was created in the fake client
	workspaceList := &workspacev1alpha1.WorkspaceList{}
	err = client.List(context.Background(), workspaceList)
	assert.NoError(t, err)

	// Validate the created workspace
	assert.Len(t, workspaceList.Items, 1)
	assert.Equal(t, "test-workspace", workspaceList.Items[0].Name)
}

func TestUpdateWorkspace(t *testing.T) {

	// Setup test client and fake client builder
	client, c := setupFakeClient()

	// Pre-create the Workspace in the fake client
	existingWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "old-namespace",
		},
	}

	err := client.Create(context.Background(), existingWorkspace)
	assert.NoError(t, err)

	// Define the workspace settings for the update
	workspaceSettings := models.WorkspaceSettings{
		Name:        "test-workspace",
		Account:     uuid.New(),
		MemberGroup: "test-group",
	}

	// Call UpdateWorkspace
	err = UpdateWorkspace(context.Background(), client, workspaceSettings, c)
	assert.NoError(t, err)

	// Retrieve and verify the updated workspace
	updatedWorkspace := &workspacev1alpha1.Workspace{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace",
		Namespace: "workspaces",
	}, updatedWorkspace)
	assert.NoError(t, err)

	// Assert that the updated namespace matches the expected value
	assert.Equal(t, "ws-test-workspace", updatedWorkspace.Spec.Namespace)
}

func TestDeleteWorkspace(t *testing.T) {
	// Setup test client and fake client builder
	client, _ := setupFakeClient()

	// Pre-create the Workspace in the fake client
	existingWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
	}

	err := client.Create(context.Background(), existingWorkspace)
	assert.NoError(t, err)

	// Define the workspace settings for deletion
	workspaceSettings := models.WorkspaceSettings{
		Name: "test-workspace",
	}

	// Call DeleteWorkspace
	err = DeleteWorkspace(context.Background(), client, workspaceSettings)
	assert.NoError(t, err)

	// Attempt to retrieve the deleted workspace
	deletedWorkspace := &workspacev1alpha1.Workspace{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace",
		Namespace: "workspaces",
	}, deletedWorkspace)

	// Verify that the workspace no longer exists
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestBuildWorkspace(t *testing.T) {

	// Setup test client and fake client builder
	_, c := setupFakeClient()

	// Mock WorkspaceSettings
	workspaceSettings := models.WorkspaceSettings{
		Name:        "test-workspace",
		Account:     uuid.New(),
		MemberGroup: "test-workspace",
		Stores: &[]models.Stores{
			{
				Object: []models.ObjectStore{
					{Name: "test-workspace-object-store"},
				},
				Block: []models.BlockStore{
					{Name: "test-workspace-block-store"},
				},
			},
		},
	}

	// Call buildWorkspace
	result := buildWorkspace(workspaceSettings, c)

	// Validate the generated Workspace object
	assert.Equal(t, "test-workspace", result.Name)
	assert.Equal(t, "ws-test-workspace", result.Spec.Namespace)
	assert.Len(t, result.Spec.AWS.S3.Buckets, 1)
	assert.Equal(t, c.AWS.Bucket, result.Spec.AWS.S3.Buckets[0].Name)
	assert.Len(t, result.Spec.AWS.EFS.AccessPoints, 1)
	assert.Equal(t, "test-workspace-block-store", result.Spec.AWS.EFS.AccessPoints[0].Name)
	assert.Len(t, result.Spec.Storage.PersistentVolumes, 1)
	assert.Equal(t, "pv-test-workspace-block-store", result.Spec.Storage.PersistentVolumes[0].Name)
}
