package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeToken(t *testing.T, token string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "TOKEN")
	require.NoError(t, os.WriteFile(path, []byte(token), 0o600))
	return path
}

func TestAuthenticationAnonymous(t *testing.T) {
	auth, err := PulsarConfig{}.Authentication()
	assert.NoError(t, err)
	assert.Nil(t, auth)
}

func TestAuthenticationMissingTokenFile(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "TOKEN")
	_, err := PulsarConfig{TokenFile: missing}.Authentication()
	assert.ErrorContains(t, err, "pulsar tokenFile is set but cannot be read")
	assert.ErrorContains(t, err, missing)
}

func TestAuthenticationTokenFile(t *testing.T) {
	auth, err := PulsarConfig{TokenFile: writeToken(t, "header.payload.signature\n")}.Authentication()
	require.NoError(t, err)
	require.NotNil(t, auth)

	// The client reads the token when it is created, without connecting.
	client, err := pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://localhost:6650", Authentication: auth})
	require.NoError(t, err)
	client.Close()
}

func TestAuthenticationEmptyTokenFile(t *testing.T) {
	auth, err := PulsarConfig{TokenFile: writeToken(t, "")}.Authentication()
	require.NoError(t, err)

	_, err = pulsar.NewClient(pulsar.ClientOptions{URL: "pulsar://localhost:6650", Authentication: auth})
	assert.Error(t, err)
}
