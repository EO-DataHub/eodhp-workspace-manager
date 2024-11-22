package k8s

import (
	"context"
	"net/http"
	"testing"
	"time"

	workspacev1alpha1 "github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestProcessWorkspace(t *testing.T) {

	// Setup a fake client and configuration
	client, config := setupFakeClient()
	ctx := context.Background()

	workspaceSettings := models.WorkspaceSettings{
		Name:        "test-workspace",
		Account:     uuid.New(),
		MemberGroup: "test-group",
	}

	// Test "creating" case
	workspaceSettings.Status = "creating"
	err := ProcessWorkspace(ctx, client, config, workspaceSettings)
	assert.NoError(t, err)

	// Verify that the workspace was created
	createdWorkspace := &workspacev1alpha1.Workspace{}
	err = client.Get(ctx, types.NamespacedName{Name: "test-workspace", Namespace: "workspaces"}, createdWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "test-workspace", createdWorkspace.Name)
	assert.Equal(t, "ws-test-workspace", createdWorkspace.Spec.Namespace)

	// Test "updating" case
	workspaceSettings.Status = "updating"
	err = ProcessWorkspace(ctx, client, config, workspaceSettings)
	assert.NoError(t, err)

	// Verify that the workspace was updated
	updatedWorkspace := &workspacev1alpha1.Workspace{}
	err = client.Get(ctx, types.NamespacedName{Name: "test-workspace", Namespace: "workspaces"}, updatedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, "ws-test-workspace", updatedWorkspace.Spec.Namespace)

	// Test "deleting" case
	workspaceSettings.Status = "deleting"
	err = ProcessWorkspace(ctx, client, config, workspaceSettings)
	assert.NoError(t, err)

	// Verify that the workspace was deleted
	deletedWorkspace := &workspacev1alpha1.Workspace{}
	err = client.Get(ctx, types.NamespacedName{Name: "test-workspace", Namespace: "workspaces"}, deletedWorkspace)
	assert.Error(t, err) // Should return an error because the workspace no longer exists

	// Test "unknown" status
	workspaceSettings.Status = "unknown"
	err = ProcessWorkspace(ctx, client, config, workspaceSettings)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown status")
}

func TestListenForWorkspaceStatusUpdates(t *testing.T) {

	// Create a runtime scheme and register the Workspace CRD
	scheme := runtime.NewScheme()
	assert.NoError(t, workspacev1alpha1.AddToScheme(scheme))

	fakeConfig := &rest.Config{}

	// Initialize a manager
	mgr, err := ctrl.NewManager(fakeConfig, ctrl.Options{
		Scheme: scheme,
	})
	assert.NoError(t, err)

	// Create a channel to receive status updates
	statusUpdates := make(chan models.WorkspaceStatus, 10)

	// Context with timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Start the informer listener in a goroutine
	go func() {
		err := ListenForWorkspaceStatusUpdates(ctx, mgr, statusUpdates)
		assert.NoError(t, err)
	}()

	// Simulate an UpdateFunc call
	fakeOldWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "old-namespace",
		},
	}
	fakeNewWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "new-namespace",
		},
	}

	// Directly trigger the UpdateFunc for testing
	handleUpdate(fakeOldWorkspace, fakeNewWorkspace, statusUpdates)

	// Wait for the status update to appear in the channel
	var receivedStatus models.WorkspaceStatus
	select {
	case receivedStatus = <-statusUpdates:
		// Verify the status update
		assert.Equal(t, "test-workspace", receivedStatus.Name)
		assert.Equal(t, "new-namespace", receivedStatus.Namespace)
	case <-ctx.Done():
		t.Fatal("Timed out waiting for status update")
	}
}

func TestHandleUpdate_InvalidInputs(t *testing.T) {
	statusUpdates := make(chan models.WorkspaceStatus, 10)

	// Pass invalid objects to the handler
	handleUpdate(nil, nil, statusUpdates)
	handleUpdate("invalid", "invalid", statusUpdates)

	select {
	case <-statusUpdates:
		t.Fatal("Expected no status updates for invalid inputs")
	default:
		// No updates as expected
	}
}

func TestListenForWorkspaceStatusUpdates_UpdateHandlerInvoked(t *testing.T) {

	// Create a manager and register the Workspace CRD
	scheme := runtime.NewScheme()
	assert.NoError(t, workspacev1alpha1.AddToScheme(scheme))

	fakeConfig := &rest.Config{}

	// Create a custom RESTMapper
	gvk := workspacev1alpha1.GroupVersion.WithKind("Workspace")
	restMapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{workspacev1alpha1.GroupVersion})
	restMapper.Add(gvk, meta.RESTScopeNamespace)

	mgr, err := ctrl.NewManager(fakeConfig, ctrl.Options{
		Scheme: scheme,
		MapperProvider: func(c *rest.Config, httpClient *http.Client) (meta.RESTMapper, error) {
			return restMapper, nil
		},
	})
	assert.NoError(t, err)

	// Create a channel to capture status updates
	statusUpdates := make(chan models.WorkspaceStatus, 10)

	// Context with timeout to avoid hanging tests
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start the listener in a goroutine
	go func() {
		err := ListenForWorkspaceStatusUpdates(ctx, mgr, statusUpdates)
		assert.NoError(t, err)
	}()

	// Simulate an UpdateFunc being invoked
	fakeOldWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "old-namespace",
		},
	}
	fakeNewWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "new-namespace",
		},
	}

	// Trigger the UpdateFunc manually
	informer, err := mgr.GetCache().GetInformer(ctx, &workspacev1alpha1.Workspace{})
	assert.NoError(t, err)
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			handleUpdate(oldObj, newObj, statusUpdates)
		},
	})

	// Manually invoke the UpdateFunc
	handleUpdate(fakeOldWorkspace, fakeNewWorkspace, statusUpdates)

	// Verify the status update is sent to the channel
	select {
	case update := <-statusUpdates:
		assert.Equal(t, "test-workspace", update.Name)
		assert.Equal(t, "new-namespace", update.Namespace)
	default:
		t.Fatal("Expected status update but did not receive any")
	}
}
func TestHandleUpdate(t *testing.T) {

	// Create a channel to capture status updates
	statusUpdates := make(chan models.WorkspaceStatus, 1)

	// Case where status has changed
	oldWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "old-namespace",
		},
	}
	newWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-workspace",
		},
		Status: workspacev1alpha1.WorkspaceStatus{
			Namespace: "new-namespace",
		},
	}

	handleUpdate(oldWorkspace, newWorkspace, statusUpdates)

	// Verify the status update is sent to the channel
	select {
	case update := <-statusUpdates:
		assert.Equal(t, "test-workspace", update.Name)
		assert.Equal(t, "new-namespace", update.Namespace)
	default:
		t.Fatal("Expected status update but did not receive any")
	}

	// Case where status has not changed
	oldWorkspace.Status = newWorkspace.Status

	handleUpdate(oldWorkspace, newWorkspace, statusUpdates)

	select {
	case <-statusUpdates:
		t.Fatal("Did not expect a status update when status has not changed")
	default:
		// Expected behavior, no update sent
	}

	// Case where oldObj cannot be cast to Workspace
	handleUpdate("invalid", newWorkspace, statusUpdates)
	// No panic expected

	// Case where newObj cannot be cast to Workspace
	handleUpdate(oldWorkspace, "invalid", statusUpdates)
	// No panic expected

	// Case where channel is full
	statusUpdates = make(chan models.WorkspaceStatus, 1)
	statusUpdates <- models.WorkspaceStatus{} // Fill the channel

	handleUpdate(oldWorkspace, newWorkspace, statusUpdates)
	// Should log a warning and not block

	// Ensure channel still has only one item
	select {
	case <-statusUpdates:
		// Consumed the existing item
	default:
		t.Fatal("Expected status update in channel")
	}

	// Ensure no additional items in the channel
	select {
	case <-statusUpdates:
		t.Fatal("Did not expect more than one status update in channel")
	default:
		// No more items as expected
	}
}
