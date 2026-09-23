package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const pulsarConfigTemplate = `pulsar:
  url: pulsar://pulsar-proxy.pulsar:6650
  topicProducer: persistent://public/default/workspace-status
  topicConsumer: {{ .TEST_TOPIC_CONSUMER }}
  subscription: workspace-manager-sub
  tokenFile: "{{ .PULSAR_TOKEN_FILE }}"
`

func TestLoadConfigPulsar(t *testing.T) {
	t.Setenv("TEST_TOPIC_CONSUMER", "persistent://public/default/workspace-settings,persistent://public/workspaces/workspace-settings")
	t.Setenv("PULSAR_TOKEN_FILE", "/var/run/secrets/pulsar-token/TOKEN")

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(pulsarConfigTemplate), 0o600))

	cfg := LoadConfig(path)
	assert.Equal(t, "/var/run/secrets/pulsar-token/TOKEN", cfg.Pulsar.TokenFile)
	assert.Equal(t, []string{
		"persistent://public/default/workspace-settings",
		"persistent://public/workspaces/workspace-settings",
	}, cfg.Pulsar.ConsumerTopics())
}

func TestConsumerTopics(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{name: "empty", value: "", want: nil},
		{name: "single", value: "persistent://public/default/a", want: []string{"persistent://public/default/a"}},
		{
			name:  "list with spaces",
			value: "persistent://public/default/a, persistent://public/workspaces/a",
			want:  []string{"persistent://public/default/a", "persistent://public/workspaces/a"},
		},
		{
			name:  "empty entries",
			value: ",persistent://public/default/a,,persistent://public/workspaces/a,",
			want:  []string{"persistent://public/default/a", "persistent://public/workspaces/a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PulsarConfig{TopicConsumer: tt.value}.ConsumerTopics())
		})
	}
}
