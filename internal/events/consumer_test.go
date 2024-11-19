package events

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/EO-DataHub/eodhp-workspace-manager/internal/utils"
	"github.com/EO-DataHub/eodhp-workspace-manager/models"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/require"
)

type MockWorkspaceOperator struct {
	ProcessedMessages []models.WorkspaceSettings
	FailProcessing    bool
}

func (m *MockWorkspaceOperator) ProcessMessage(ctx context.Context, payload models.WorkspaceSettings) error {
	if m.FailProcessing {
		return errors.New("simulated processing error")
	}
	m.ProcessedMessages = append(m.ProcessedMessages, payload)
	return nil
}

func TestConfigurationConsumer(t *testing.T) {
	// Get Pulsar URL from environment
	pulsarURL := "pulsar://localhost:6650"
	require.NotEmpty(t, pulsarURL, "PULSAR_URL must be set for the test")

	// Set up Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: pulsarURL,
	})
	require.NoError(t, err)
	defer client.Close()

	// Define a test topic and subscription
	topic := "test-topic"
	subscription := "test-subscription"

	// Set up a mock operator
	mockOperator := &MockWorkspaceOperator{}

	// Create a test configuration
	config := &utils.Config{
		Pulsar: utils.PulsarConfig{
			TopicConsumer: topic,
			Subscription:  subscription,
		},
	}

	// Create the consumer
	consumer := NewConfigurationConsumer(client, mockOperator, config)
	defer consumer.Stop()

	// Start the consumer
	go consumer.Start()

	// Set up a producer to send messages
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	require.NoError(t, err)
	defer producer.Close()

	// Send a test message
	testPayload := models.WorkspaceSettings{Name: "TestWorkspace"}
	message, err := json.Marshal(testPayload)
	require.NoError(t, err)

	_, err = producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: message,
	})
	require.NoError(t, err)

	// Wait for the consumer to process the message
	time.Sleep(2 * time.Second)

	// Verify the message was processed
	require.Len(t, mockOperator.ProcessedMessages, 1)
	require.Equal(t, "TestWorkspace", mockOperator.ProcessedMessages[0].Name)
}
