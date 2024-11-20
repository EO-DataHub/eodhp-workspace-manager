package manager

import (
	"context"
	"testing"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/events"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/stretchr/testify/assert"
)

// MockK8sInterface is a mock implementation of K8sInterface
type MockK8sInterface struct {
	CreateWorkspaceFunc func(ctx context.Context, payload models.WorkspaceSettings) error
	UpdateWorkspaceFunc func(ctx context.Context, payload models.WorkspaceSettings) error
	DeleteWorkspaceFunc func(ctx context.Context, payload models.WorkspaceSettings) error
}

// CreateWorkspace creates a new workspace
func (m *MockK8sInterface) CreateWorkspace(ctx context.Context, payload models.WorkspaceSettings) error {
	if m.CreateWorkspaceFunc != nil {
		return m.CreateWorkspaceFunc(ctx, payload)
	}
	return nil
}

// UpdateWorkspace updates an existing workspace
func (m *MockK8sInterface) UpdateWorkspace(ctx context.Context, payload models.WorkspaceSettings) error {
	if m.UpdateWorkspaceFunc != nil {
		return m.UpdateWorkspaceFunc(ctx, payload)
	}
	return nil
}

// DeleteWorkspace deletes an existing workspace
func (m *MockK8sInterface) DeleteWorkspace(ctx context.Context, payload models.WorkspaceSettings) error {
	if m.DeleteWorkspaceFunc != nil {
		return m.DeleteWorkspaceFunc(ctx, payload)
	}
	return nil
}

// MockPulsarClient is a mock implementation of PulsarClient
type MockPulsarClient struct{}

func TestProcessMessage_Creating(t *testing.T) {

	// Mock K8sInterface with a CreateWorkspace function
	mockK8s := &MockK8sInterface{
		CreateWorkspaceFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			assert.Equal(t, "creating", payload.Status)
			return nil
		},
	}

	// Mock PulsarClient
	mockPulsar := &events.MockPulsarClient{}

	// Create a new WorkspaceOperator
	operator := NewWorkspaceOperator(mockK8s, mockPulsar)

	// Simulate a "creating" message
	payload := models.WorkspaceSettings{
		Name:   "test-workspace",
		Status: "creating",
	}

	// Process the message
	err := operator.ProcessMessage(context.Background(), payload)
	assert.NoError(t, err)
}

func TestProcessMessage_Updating(t *testing.T) {

	// Mock K8sInterface with an UpdateWorkspace function
	mockK8s := &MockK8sInterface{
		UpdateWorkspaceFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			assert.Equal(t, "updating", payload.Status)
			return nil
		},
	}

	// Mock PulsarClient
	mockPulsar := &events.MockPulsarClient{}

	// Create a new WorkspaceOperator
	operator := NewWorkspaceOperator(mockK8s, mockPulsar)

	// Simulate an "updating" message
	payload := models.WorkspaceSettings{
		Name:   "test-workspace",
		Status: "updating",
	}

	// Process the message
	err := operator.ProcessMessage(context.Background(), payload)
	assert.NoError(t, err)
}

func TestProcessMessage_Deleting(t *testing.T) {

	// Mock K8sInterface with a DeleteWorkspace function
	mockK8s := &MockK8sInterface{
		DeleteWorkspaceFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			assert.Equal(t, "deleting", payload.Status)
			return nil
		},
	}

	// Mock PulsarClient
	mockPulsar := &events.MockPulsarClient{}

	// Create a new WorkspaceOperator
	operator := NewWorkspaceOperator(mockK8s, mockPulsar)

	// Simulate a "delete" message
	payload := models.WorkspaceSettings{
		Name:   "test-workspace",
		Status: "deleting",
	}

	// Process the message
	err := operator.ProcessMessage(context.Background(), payload)
	assert.NoError(t, err)
}

func TestProcessMessage_UnknownStatus(t *testing.T) {

	// Mock K8sInterface
	mockK8s := &MockK8sInterface{}

	// Mock PulsarClient
	mockPulsar := &events.MockPulsarClient{}

	// Create a new WorkspaceOperator
	operator := NewWorkspaceOperator(mockK8s, mockPulsar)

	// Simulate an unknown status
	payload := models.WorkspaceSettings{
		Name:   "test-workspace",
		Status: "unknown-status",
	}

	// Process the message
	err := operator.ProcessMessage(context.Background(), payload)
	assert.NoError(t, err)
}
