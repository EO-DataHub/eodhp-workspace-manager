package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/EO-DataHub/eodhp-workspace-controller/api/v1alpha1"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// MockPulsarProducer is a mock implementation of the Pulsar Producer interface
type MockPulsarProducer struct {
	pulsar.Producer

	SendFunc  func(ctx context.Context, msg *pulsar.ProducerMessage) (pulsar.MessageID, error)
	CloseFunc func()
}

// Send sends a message to the Pulsar topic
func (m *MockPulsarProducer) Send(ctx context.Context, msg *pulsar.ProducerMessage) (pulsar.MessageID, error) {
	if m.SendFunc != nil {
		return m.SendFunc(ctx, msg)
	}
	return nil, nil
}

// Close closes the producer
func (m *MockPulsarProducer) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

func TestNewStatusPublisher(t *testing.T) {

	// Mock Pulsar client
	mockPulsarClient := &MockPulsarClient{}

	// Create a fake K8s client
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	mockK8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	// Create a new StatusPublisher
	publisher := NewStatusPublisher(mockK8sClient, mockPulsarClient, "test-topic")

	// Assertions to verify correct initialization
	assert.NotNil(t, publisher)
	assert.Equal(t, mockPulsarClient, publisher.pulsarClient)
	assert.Equal(t, mockK8sClient, publisher.k8sClient)
	assert.Equal(t, "test-topic", publisher.topic)
}

func TestHandleUpdate_NoStatusChange(t *testing.T) {

	// Simulate two Workspaces with identical statuses
	oldWorkspace := &v1alpha1.Workspace{
		Status: v1alpha1.WorkspaceStatus{
			Namespace: "workspaces",
			AWS: v1alpha1.AWSStatus{
				EFS: v1alpha1.EFSStatus{
					AccessPoints: []v1alpha1.EFSAccessStatus{
						{Name: "AccessPoint1", FSID: "fs-123"},
					},
				},
				S3: v1alpha1.S3Status{
					Buckets: []v1alpha1.S3BucketStatus{
						{Name: "Bucket1", Path: "/data"},
					},
				},
			},
		},
	}
	newWorkspace := &v1alpha1.Workspace{
		Status: v1alpha1.WorkspaceStatus{
			Namespace: "workspaces",
			AWS: v1alpha1.AWSStatus{
				EFS: v1alpha1.EFSStatus{
					AccessPoints: []v1alpha1.EFSAccessStatus{
						{Name: "AccessPoint1", FSID: "fs-123"},
					},
				},
				S3: v1alpha1.S3Status{
					Buckets: []v1alpha1.S3BucketStatus{
						{Name: "Bucket1", Path: "/data"},
					},
				},
			},
		},
	}

	// Mock Pulsar client (not expected to send any messages)
	mockPulsarClient := &MockPulsarClient{}

	// Initialize the StatusPublisher
	publisher := &StatusPublisher{
		pulsarClient: mockPulsarClient,
		topic:        "test-topic",
	}

	// Call handleUpdate with identical old and new Workspace states
	publisher.handleUpdate(oldWorkspace, newWorkspace)

}

func TestHandleUpdate_DetectStatusChange(t *testing.T) {

	// Simulate the old state of the Workspace
	oldWorkspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Status: v1alpha1.WorkspaceStatus{
			Namespace: "Creating",
		},
	}

	// Simulate the new state of the Workspace with a changed status
	newWorkspace := &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-workspace",
			Namespace: "test-namespace",
		},
		Status: v1alpha1.WorkspaceStatus{
			Namespace: "Created",
			AWS: v1alpha1.AWSStatus{
				EFS: v1alpha1.EFSStatus{
					AccessPoints: []v1alpha1.EFSAccessStatus{
						{Name: "AccessPoint2", FSID: "fs-456"},
					},
				},
				S3: v1alpha1.S3Status{
					Buckets: []v1alpha1.S3BucketStatus{
						{Name: "Bucket2", Path: "/data/updated"},
					},
				},
				Role: v1alpha1.AWSRoleStatus{
					Name: "workspace-role",
					ARN:  "arn:aws:iam::123456789012:role/workspace-role",
				},
			},
		},
	}

	// Mock Pulsar producer to capture the payload sent
	mockProducer := &MockPulsarProducer{
		SendFunc: func(ctx context.Context, msg *pulsar.ProducerMessage) (pulsar.MessageID, error) {

			// Deserialize the actual payload sent
			var actualPayload map[string]interface{}
			err := json.Unmarshal(msg.Payload, &actualPayload)
			assert.NoError(t, err)

			// Expected payload based on the newWorkspace's status
			expectedPayload := map[string]interface{}{
				"workspaceName": "test-workspace",
				"namespace":     "test-namespace",
				"status": map[string]interface{}{
					"namespace": "Created",
					"aws": map[string]interface{}{
						"efs": map[string]interface{}{
							"accessPoints": []interface{}{
								map[string]interface{}{
									"name": "AccessPoint2",
									"fsID": "fs-456",
								},
							},
						},
						"s3": map[string]interface{}{
							"buckets": []interface{}{
								map[string]interface{}{
									"name": "Bucket2",
									"path": "/data/updated",
								},
							},
						},
						"role": map[string]interface{}{
							"name": "workspace-role",
							"arn":  "arn:aws:iam::123456789012:role/workspace-role",
						},
					},
				},
			}

			// Validate the actual payload matches the expected payload
			assert.Equal(t, expectedPayload, actualPayload)

			return nil, nil
		},
	}

	// Mock Pulsar client to return the mock producer
	mockPulsarClient := &MockPulsarClient{
		CreateProducerFunc: func(options pulsar.ProducerOptions) (pulsar.Producer, error) {
			return mockProducer, nil
		},
	}

	// Initialize the StatusPublisher
	publisher := &StatusPublisher{
		pulsarClient: mockPulsarClient,
		topic:        "test-topic",
	}

	// Call handleUpdate directly to simulate the status change detection
	publisher.handleUpdate(oldWorkspace, newWorkspace)
}

func TestSendStatusUpdate_ProducerError(t *testing.T) {

	// Mock Pulsar client to simulate producer creation failure
	mockPulsarClient := &MockPulsarClient{
		CreateProducerFunc: func(options pulsar.ProducerOptions) (pulsar.Producer, error) {
			return nil, assert.AnError // Simulate an error
		},
	}

	// Initialize the StatusPublisher
	publisher := &StatusPublisher{
		pulsarClient: mockPulsarClient,
		topic:        "test-topic",
	}

	// Call sendStatusUpdate with arbitrary payload
	err := publisher.sendStatusUpdate([]byte(`{"key":"value"}`))

	// Assert that an error is returned
	assert.Error(t, err)
}
