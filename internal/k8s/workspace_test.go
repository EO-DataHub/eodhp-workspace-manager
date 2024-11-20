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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupTestClient() (*K8sClient, *fake.ClientBuilder) {
	// Define the updated configuration
	testConfig := &utils.Config{
		LogLevel: "debug",
		Pulsar: utils.PulsarConfig{
			URL:           "pulsar://localhost:6650",
			TopicProducer: "test-producer-topic",
			TopicConsumer: "test-consumer-topic",
			Subscription:  "test-subscription",
		},
		AWS: utils.AWSConfig{
			Cluster: "test-cluster",
			FSID:    "fs-test",
		},
		Storage: utils.StorageConfig{
			Size:         "20Gi",
			StorageClass: "gp2",
			PVCName:      "workspace-pvc",
			Driver:       "efs.csi.aws.com",
		},
	}

	// Create a fake Kubernetes client
	scheme := runtime.NewScheme()
	_ = workspacev1alpha1.AddToScheme(scheme)

	fakeClientBuilder := fake.NewClientBuilder().WithScheme(scheme)
	fakeClient := fakeClientBuilder.Build()

	// Initialize the k8s Client
	k8sClient := &K8sClient{
		client: fakeClient,
		Config: testConfig,
	}

	return k8sClient, fakeClientBuilder
}

func TestMapObjectStoresToS3Buckets(t *testing.T) {

	// Setup test client and fake client builder
	client, _ := setupTestClient()

	objectStores := []models.ObjectStore{
		{Name: "obj1", Path: "/data/obj1", EnvVar: "OBJ1_ENV"},
		{Name: "obj2", Path: "/data/obj2", EnvVar: "OBJ2_ENV"},
	}

	// Call MapObjectStoresToS3Buckets
	result := client.MapObjectStoresToS3Buckets("test-workspace", objectStores)

	assert.Len(t, result, 2)
	assert.Equal(t, "obj1", result[0].Name)
	assert.Equal(t, "/data/obj1", result[0].Path)
	assert.Equal(t, "OBJ1_ENV", result[0].EnvVar)
	assert.Equal(t, "test-cluster-test-workspace-s3", result[0].AccessPointName)
}

func TestMapBlockStoresToEFSAccessPoints(t *testing.T) {
	
	// Setup test client and fake client builder
	client, _ := setupTestClient()

	blockStores := []models.BlockStore{
		{Name: "block1"},
		{Name: "block2"},
	}

	// Call MapBlockStoresToEFSAccessPoints
	result := client.MapBlockStoresToEFSAccessPoints("test-workspace", blockStores)

	assert.Len(t, result, 2)
	assert.Equal(t, "block1", result[0].Name)
	assert.Equal(t, "fs-test", result[0].FSID)
	assert.Equal(t, "/workspaces/test-workspace", result[0].RootDirectory)
	assert.Equal(t, "0755", result[0].Permissions)
	assert.Equal(t, int64(1000), result[0].User.UID)
}

func TestGenerateStorageConfig(t *testing.T) {

	// Setup test client and fake client builder
	client, _ := setupTestClient()

	// Call GenerateStorageConfig
	result := client.GenerateStorageConfig("test-workspace")

	assert.Len(t, result.PersistentVolumes, 1)
	assert.Equal(t, "pv-test-workspace-workspace", result.PersistentVolumes[0].Name)
	assert.Equal(t, "gp2", result.PersistentVolumes[0].StorageClass)
	assert.Equal(t, "20Gi", result.PersistentVolumes[0].Size)
	assert.Equal(t, "efs.csi.aws.com", result.PersistentVolumes[0].VolumeSource.Driver)
}

func TestCreateWorkspace(t *testing.T) {

	// Setup test client and fake client builder
	client, fakeClientBuilder := setupTestClient()

	// Build the fake client once and assign it to the Client instance
	fakeClient := fakeClientBuilder.Build()
	client.client = fakeClient

	workspaceSettings := models.WorkspaceSettings{
		Name:        "test-workspace",
		Account:     uuid.New(),
		MemberGroup: "test-group",
	}

	// Call CreateWorkspace
	err := client.CreateWorkspace(context.Background(), workspaceSettings)
	assert.NoError(t, err)

	// Verify the workspace was created in the fake client
	workspaceList := &workspacev1alpha1.WorkspaceList{}
	err = fakeClient.List(context.Background(), workspaceList)
	assert.NoError(t, err)

	// Validate the created workspace
	assert.Len(t, workspaceList.Items, 1)
	assert.Equal(t, "test-workspace", workspaceList.Items[0].Name)
}

func TestUpdateWorkspace(t *testing.T) {
	
	// Setup test client and fake client builder
	client, fakeClientBuilder := setupTestClient()

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

	// Build the fake client and add the existing workspace
	fakeClient := fakeClientBuilder.Build()
	err := fakeClient.Create(context.Background(), existingWorkspace)
	assert.NoError(t, err)

	// Define the workspace settings for the update
	workspaceSettings := models.WorkspaceSettings{
		Name:        "test-workspace",
		Account:     uuid.New(),
		MemberGroup: "test-group",
	}

	// Assign the fake client to the Client struct
	client.client = fakeClient

	// Call UpdateWorkspace
	err = client.UpdateWorkspace(context.Background(), workspaceSettings)
	assert.NoError(t, err)

	// Retrieve and verify the updated workspace
	updatedWorkspace := &workspacev1alpha1.Workspace{}
	err = fakeClient.Get(context.Background(), types.NamespacedName{
		Name:      "test-workspace",
		Namespace: "workspaces",
	}, updatedWorkspace)
	assert.NoError(t, err)

	// Assert that the updated namespace matches the expected value
	assert.Equal(t, "ws-test-workspace", updatedWorkspace.Spec.Namespace)
}
