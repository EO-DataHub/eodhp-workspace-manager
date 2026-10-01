package k8s

import (
	"context"
	"testing"

	"github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/stretchr/testify/assert"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCreateWorkspace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	ctx := context.Background()
	cfg := &utils.Config{
		AWS: utils.AWSConfig{
			Bucket:  "test-bucket",
			Cluster: "test-cluster",
			FSID:    "fs-12345",
		},
		Storage: utils.StorageConfig{
			Driver:       "efs",
			StorageClass: "standard",
			Size:         "10Gi",
		},
	}

	payload := models.WorkspaceSettings{
		Name: "test-ws",
		Stores: &[]models.Stores{
			{
				Object: []models.ObjectStore{{Name: "input"}},
				Block:  []models.BlockStore{{Name: "data"}},
			},
		},
		Status: "creating",
		Owner:  "owner",
	}

	err := CreateWorkspace(ctx, fakeClient, payload, cfg)
	assert.NoError(t, err)

	created := &v1alpha1.Workspace{}
	err = fakeClient.Get(ctx, client.ObjectKey{Name: "test-ws", Namespace: "workspaces"}, created)
	assert.NoError(t, err)
	assert.Equal(t, "test-ws", created.Name)
	assert.Equal(t, "ws-test-ws", created.Spec.Namespace)
}

func TestUpdateWorkspace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	cfg := &utils.Config{
		AWS: utils.AWSConfig{
			Bucket:  "bucket",
			Cluster: "cluster",
			FSID:    "fsid",
		},
		Storage: utils.StorageConfig{
			Driver:       "efs",
			StorageClass: "standard",
			Size:         "5Gi",
		},
	}

	initial := CreateWorkspace(ctx, fakeClient, models.WorkspaceSettings{
		Name: "update-ws",
		Stores: &[]models.Stores{
			{
				Block:  []models.BlockStore{{Name: "block1"}},
				Object: []models.ObjectStore{{Name: "object1"}},
			},
		},
	}, cfg)

	assert.NoError(t, initial)

	payload := models.WorkspaceSettings{
		Name: "update-ws",
		Stores: &[]models.Stores{
			{
				Block:  []models.BlockStore{{Name: "block2"}},
				Object: []models.ObjectStore{{Name: "object2"}},
			},
		},
		Status: "updating",
	}

	err := UpdateWorkspace(ctx, fakeClient, payload, cfg)
	assert.NoError(t, err)

	updated := &v1alpha1.Workspace{}
	err = fakeClient.Get(ctx, client.ObjectKey{Name: "update-ws", Namespace: "workspaces"}, updated)
	assert.NoError(t, err)
	assert.Equal(t, "update-ws", updated.Name)
}

func TestUpdateWorkspacePreservesControllerMetadata(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	cfg := &utils.Config{
		AWS: utils.AWSConfig{
			Bucket:  "bucket",
			Cluster: "cluster",
			FSID:    "fsid",
		},
		Storage: utils.StorageConfig{
			Driver:       "efs",
			StorageClass: "standard",
			Size:         "5Gi",
		},
	}

	settings := models.WorkspaceSettings{
		Name: "meta-ws",
		Stores: &[]models.Stores{
			{
				Block:  []models.BlockStore{{Name: "block1"}},
				Object: []models.ObjectStore{{Name: "object1"}},
			},
		},
	}
	assert.NoError(t, CreateWorkspace(ctx, fakeClient, settings, cfg))

	// Simulate the workspace-controller adding its finalizer and an annotation
	key := client.ObjectKey{Name: "meta-ws", Namespace: "workspaces"}
	existing := &v1alpha1.Workspace{}
	assert.NoError(t, fakeClient.Get(ctx, key, existing))
	existing.Finalizers = []string{"core.telespazio-uk.io/workspace-finalizer"}
	existing.Annotations = map[string]string{"example.com/note": "kept"}
	assert.NoError(t, fakeClient.Update(ctx, existing))

	t.Run("settings with an unchanged spec do not write", func(t *testing.T) {
		before := &v1alpha1.Workspace{}
		assert.NoError(t, fakeClient.Get(ctx, key, before))

		category := "academic"
		unchanged := settings
		unchanged.Status = "updating"
		unchanged.Category = &category
		assert.NoError(t, UpdateWorkspace(ctx, fakeClient, unchanged, cfg))

		after := &v1alpha1.Workspace{}
		assert.NoError(t, fakeClient.Get(ctx, key, after))
		assert.Equal(t, before.ResourceVersion, after.ResourceVersion)
	})

	t.Run("a spec change keeps finalizers and annotations", func(t *testing.T) {
		changed := settings
		changed.Status = "updating"
		changed.Stores = &[]models.Stores{
			{
				Block:  []models.BlockStore{{Name: "block2"}},
				Object: []models.ObjectStore{{Name: "object2"}},
			},
		}
		assert.NoError(t, UpdateWorkspace(ctx, fakeClient, changed, cfg))

		updated := &v1alpha1.Workspace{}
		assert.NoError(t, fakeClient.Get(ctx, key, updated))
		assert.Equal(t, []string{"core.telespazio-uk.io/workspace-finalizer"}, updated.Finalizers)
		assert.Equal(t, "kept", updated.Annotations["example.com/note"])
		assert.Equal(t, "block2", updated.Spec.AWS.EFS.AccessPoints[0].Name)
	})
}

func TestDeleteWorkspace(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	ctx := context.Background()

	// Pre-create workspace
	createErr := CreateWorkspace(ctx, fakeClient, models.WorkspaceSettings{
		Name: "delete-ws",
		Stores: &[]models.Stores{
			{
				Block:  []models.BlockStore{{Name: "block"}},
				Object: []models.ObjectStore{{Name: "object"}},
			},
		},
	}, &utils.Config{
		AWS: utils.AWSConfig{Bucket: "bucket", Cluster: "cluster", FSID: "fsid"},
		Storage: utils.StorageConfig{
			Driver:       "efs",
			StorageClass: "sc",
			Size:         "5Gi",
		},
	})
	assert.NoError(t, createErr)

	err := DeleteWorkspace(ctx, fakeClient, models.WorkspaceSettings{
		Name: "delete-ws",
	})
	assert.NoError(t, err)

	deleted := &v1alpha1.Workspace{}
	getErr := fakeClient.Get(ctx, client.ObjectKey{Name: "delete-ws", Namespace: "workspaces"}, deleted)
	assert.Error(t, getErr) // Should not find the object anymore
}

func TestUpdateWorkspaceNotFoundIsSkipped(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// An update for a workspace that doesn't exist must not fail, or Pulsar redelivers it forever
	err := UpdateWorkspace(context.Background(), fakeClient, models.WorkspaceSettings{Name: "gone-ws", Status: "updating"}, &utils.Config{})
	assert.NoError(t, err)

	err = fakeClient.Get(context.Background(), client.ObjectKey{Name: "gone-ws", Namespace: "workspaces"}, &v1alpha1.Workspace{})
	assert.True(t, apierrors.IsNotFound(err))
}
