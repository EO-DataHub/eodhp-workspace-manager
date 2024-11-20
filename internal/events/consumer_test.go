package events

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
)

// MockConsumer is a mock implementation of the Pulsar Consumer interface
type MockConsumer struct {
	pulsar.Consumer
	ReceiveFunc func(ctx context.Context) (pulsar.Message, error)
	AckFunc     func(msg pulsar.Message) error
	NackFunc    func(msg pulsar.Message)
	CloseFunc   func()
}

// Receive receives a message from the Pulsar topic
func (m *MockConsumer) Receive(ctx context.Context) (pulsar.Message, error) {
	if m.ReceiveFunc != nil {
		return m.ReceiveFunc(ctx)
	}
	return nil, nil
}

// Ack acknowledges a message
func (m *MockConsumer) Ack(msg pulsar.Message) error {
	if m.AckFunc != nil {
		return m.AckFunc(msg)
	}
	return nil
}

// Nack negatively acknowledges a message
func (m *MockConsumer) Nack(msg pulsar.Message) {
	if m.NackFunc != nil {
		m.NackFunc(msg)
	}
}

// Close closes the consumer
func (m *MockConsumer) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// MockWorkspaceOperator is a mock implementation of the WorkspaceOperator interface
type MockWorkspaceOperator struct {
	ProcessMessageFunc func(ctx context.Context, payload models.WorkspaceSettings) error
}

// ProcessMessage processes a message from a Pulsar topic
func (m *MockWorkspaceOperator) ProcessMessage(ctx context.Context, payload models.WorkspaceSettings) error {
	if m.ProcessMessageFunc != nil {
		return m.ProcessMessageFunc(ctx, payload)
	}
	return nil
}

// MockPulsarMessage is a mock implementation of the Pulsar Message interface
type MockPulsarMessage struct {
	pulsar.Message
	PayloadData []byte
}

// Payload returns the message payload
func (m *MockPulsarMessage) Payload() []byte {
	return m.PayloadData
}

func TestConfigurationConsumer_handleMessage_Success(t *testing.T) {

	// Mock Pulsar message with a valid (BASIC) WorkspaceSettings payload
	mockMessage := &MockPulsarMessage{
		PayloadData: func() []byte {
			payload, _ := json.Marshal(models.WorkspaceSettings{
				Name:   "test-workspace",
				Status: "creating",
			})
			return payload
		}(),
	}

	// Mock Consumer to verify that Ack is called and Nack is not
	mockConsumer := &MockConsumer{
		AckFunc: func(msg pulsar.Message) error {
			assert.Equal(t, mockMessage, msg)
			return nil
		},
		NackFunc: func(msg pulsar.Message) {
			assert.Fail(t, "Nack should not be called")
		},
	}

	// Mock WorkspaceOperator to validate the payload passed to ProcessMessage
	mockOperator := &MockWorkspaceOperator{
		ProcessMessageFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			return nil
		},
	}

	// Create ConfigurationConsumer instance
	consumer := &ConfigurationConsumer{
		Consumer: mockConsumer,
		Operator: mockOperator,
		Config:   &utils.Config{},
		ctx:      context.Background(),
	}

	// Call handleMessage with the mock message
	consumer.handleMessage(mockMessage)
}

func TestConfigurationConsumer_handleMessage_OperatorError(t *testing.T) {

	// Mock Pulsar message with a valid (BASIC) WorkspaceSettings payload
	mockMessage := &MockPulsarMessage{
		PayloadData: func() []byte {
			payload, _ := json.Marshal(models.WorkspaceSettings{
				Name:   "test-workspace",
				Status: "creating",
			})
			return payload
		}(),
	}

	// Mock Consumer to verify that Nack is called and Ack is not
	mockConsumer := &MockConsumer{
		AckFunc: func(msg pulsar.Message) error {
			assert.Fail(t, "Ack should not be called")
			return nil
		},
		NackFunc: func(msg pulsar.Message) {
			assert.Equal(t, mockMessage, msg)
		},
	}

	// Mock WorkspaceOperator to simulate an error during ProcessMessage
	mockOperator := &MockWorkspaceOperator{
		ProcessMessageFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			return assert.AnError
		},
	}

	// Create ConfigurationConsumer instance
	consumer := &ConfigurationConsumer{
		Consumer: mockConsumer,
		Operator: mockOperator,
		Config:   &utils.Config{},
		ctx:      context.Background(),
	}

	// Call handleMessage
	consumer.handleMessage(mockMessage)
}

func TestConfigurationConsumer_Stop(t *testing.T) {

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	consumerClosed := false

	// Mock Consumer to verify that Close is called
	mockConsumer := &MockConsumer{
		CloseFunc: func() {
			consumerClosed = true
		},
	}

	// Create ConfigurationConsumer instance
	consumer := &ConfigurationConsumer{
		Consumer: mockConsumer,
		Config:   &utils.Config{},
		ctx:      ctx,
		cancel:   cancel,
	}

	// Call Stop to close the consumer and cancel the context
	consumer.Stop()

	// Verify the consumer is closed
	assert.True(t, consumerClosed, "Consumer should be closed")

	// Verify the context is canceled
	select {
	case <-ctx.Done():
	default:
		assert.Fail(t, "Context should be canceled")
	}
}

func TestConfigurationConsumer_Start(t *testing.T) {

	// Create a context with cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Ensure the context is canceled to prevent goroutine leaks
	defer cancel()

	// Mock Pulsar message with a valid (BASIC) WorkspaceSettings payload
	mockMessage := &MockPulsarMessage{
		PayloadData: []byte(`{"name": "test-workspace", "status": "creating"}`),
	}

	// Channel to signal message processing
	messageProcessed := make(chan struct{})

	// Ensures messageProcessed is closed only once
	var ackOnce sync.Once

	// Mock Consumer to simulate receiving and acknowledging a message
	mockConsumer := &MockConsumer{
		ReceiveFunc: func(ctx context.Context) (pulsar.Message, error) {
			return mockMessage, nil
		},
		AckFunc: func(msg pulsar.Message) error {
			assert.Equal(t, mockMessage, msg)
			ackOnce.Do(func() {
				close(messageProcessed)
			})
			return nil
		},
	}

	// Mock WorkspaceOperator to validate the payload passed to ProcessMessage
	mockOperator := &MockWorkspaceOperator{
		ProcessMessageFunc: func(ctx context.Context, payload models.WorkspaceSettings) error {
			assert.Equal(t, "test-workspace", payload.Name)
			return nil
		},
	}

	// Create ConfigurationConsumer instance
	consumer := &ConfigurationConsumer{
		Consumer: mockConsumer,
		Operator: mockOperator,
		Config:   &utils.Config{},
		ctx:      ctx,
		cancel:   cancel,
	}

	// Run Start in a separate goroutine
	go consumer.Start()

	// Wait for the message to be processed or timeout
	select {
	case <-messageProcessed:
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for message to be processed")
	}

	// Ensure the context cancellation stops the consumer
	cancel()
}

func TestConfigurationConsumer_handleMessage_InvalidPayload(t *testing.T) {

	// Mock Pulsar message with an invalid JSON payload
	mockMessage := &MockPulsarMessage{
		PayloadData: []byte(`invalid-json`), // Invalid JSON payload
	}

	// Mock Consumer to verify that Nack is called
	mockConsumer := &MockConsumer{
		NackFunc: func(msg pulsar.Message) {
			assert.Equal(t, mockMessage, msg)
		},
	}

	// Mock WorkspaceOperator to simulate an error during ProcessMessage
	mockOperator := &MockWorkspaceOperator{}

	// Create ConfigurationConsumer instance
	consumer := &ConfigurationConsumer{
		Consumer: mockConsumer,
		Operator: mockOperator,
		Config:   &utils.Config{},
		ctx:      context.Background(),
	}

	// Call handleMessage
	consumer.handleMessage(mockMessage)
}
