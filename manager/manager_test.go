package manager

import (
	"context"
	"os"
	"testing"

	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	workspacev1alpha1 "github.com/UKEODHP/workspace-controller/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func stringPointer(s string) *string {
	return &s
}

// Runs setup code once for all tests
func TestMain(m *testing.M) {
	// Add workspace CRD to the scheme once
	err := workspacev1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		os.Exit(1)
	}

	// Run the tests
	code := m.Run()

	os.Exit(code)
}

// Helper function to create a new fake client and manager in all the tests
func setupFakeClientAndManager(objects ...client.Object) (client.Client, *Manager) {

	// Create a new fake client with any objects passed in
	fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(objects...).Build()

	// Initialize the manager with the fake client
	mgr := NewManager(fakeClient, nil) // nil for Pulsar client as we are not testing this functionality

	return fakeClient, mgr
}

func TestCreateWorkspace(t *testing.T) {

	fakeClient, mgr := setupFakeClientAndManager()

	workspacePayload := models.WorkspacePayload{
		Name:               "test-workspace",
		CRNamespace:        "workspaces",
		TargetNamespace:    "ws-test",
		AWSRoleName:        stringPointer("aws-role"),
		ServiceAccountName: stringPointer("default"),
	}

	// Create the workspace with fake client
	err := mgr.createWorkspace(&workspacePayload)

	// Assert that no error occurred
	require.NoError(t, err)

	// Verify the workspace was created in the fake client
	createdWorkspace := &workspacev1alpha1.Workspace{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "test-workspace", Namespace: "workspaces"}, createdWorkspace)
	require.NoError(t, err)
	require.Equal(t, "ws-test", createdWorkspace.Spec.Namespace)
}

func TestUpdateWorkspace(t *testing.T) {

	existingWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Spec: workspacev1alpha1.WorkspaceSpec{
			Namespace: "ws-old",
		},
	}

	// Create a new fake client with an existing workspace
	fakeClient, mgr := setupFakeClientAndManager(existingWorkspace)

	workspacePayload := models.WorkspacePayload{
		Name:               "test-workspace",
		CRNamespace:        "workspaces",
		TargetNamespace:    "ws-updated",
		AWSRoleName:        stringPointer("aws-role-updated"),
		ServiceAccountName: stringPointer("default"),
	}

	// Update the workspace with fake client
	err := mgr.updateWorkspace(&workspacePayload)
	require.NoError(t, err)

	// Verify the workspace was updated in the fake client
	updatedWorkspace := &workspacev1alpha1.Workspace{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "test-workspace", Namespace: "workspaces"}, updatedWorkspace)
	require.NoError(t, err)
	require.Equal(t, "ws-updated", updatedWorkspace.Spec.Namespace)
	require.Equal(t, "aws-role-updated", updatedWorkspace.Spec.AWS.RoleName)
}

func TestDeleteWorkspace(t *testing.T) {

	existingWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
	}

	// Create a new fake client with an existing workspace
	fakeClient, mgr := setupFakeClientAndManager(existingWorkspace)

	workspacePayload := models.WorkspacePayload{
		Name:        "test-workspace",
		CRNamespace: "workspaces",
	}

	// Delete the workspace with fake client
	err := mgr.deleteWorkspace(&workspacePayload)
	require.NoError(t, err)

	// Verify the workspace was deleted in the fake client
	deletedWorkspace := &workspacev1alpha1.Workspace{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "test-workspace", Namespace: "workspaces"}, deletedWorkspace)
	require.Error(t, err) // Expecting an error because the workspace should no longer exist
}

func TestPatchWorkspace(t *testing.T) {

	existingWorkspace := &workspacev1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "workspaces",
		},
		Spec: workspacev1alpha1.WorkspaceSpec{
			Namespace: "ws-old",
			AWS: workspacev1alpha1.AWSSpec{
				RoleName: "aws-role-old",
			},
		},
	}

	// Create a new fake client with an existing workspace
	fakeClient, mgr := setupFakeClientAndManager(existingWorkspace)

	patchFields := map[string]interface{}{
		"spec": map[string]interface{}{
			"aws": map[string]interface{}{
				"roleName": "aws-role-patched",
			},
		},
	}
	workspacePayload := models.WorkspacePayload{
		Name:        "test-workspace",
		CRNamespace: "workspaces",
		PatchFields: patchFields,
	}

	// Patch the workspace with fake client
	err := mgr.patchWorkspace(&workspacePayload)
	require.NoError(t, err)

	// Verify the patch was applied
	patchedWorkspace := &workspacev1alpha1.Workspace{}
	err = fakeClient.Get(context.TODO(), client.ObjectKey{Name: "test-workspace", Namespace: "workspaces"}, patchedWorkspace)
	require.NoError(t, err)
	require.Equal(t, "aws-role-patched", patchedWorkspace.Spec.AWS.RoleName)
}
