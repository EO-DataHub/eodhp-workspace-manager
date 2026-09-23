package utils

import (
	"fmt"
	"os"

	"github.com/apache/pulsar-client-go/pulsar"
)

// Authentication returns token authentication read from TokenFile, or nil
// (anonymous) when TokenFile is empty. The client re-reads the file on every
// connection, so a rotated token is picked up without a restart.
func (p PulsarConfig) Authentication() (pulsar.Authentication, error) {
	if p.TokenFile == "" {
		return nil, nil
	}
	if _, err := os.Stat(p.TokenFile); err != nil {
		return nil, fmt.Errorf("pulsar tokenFile is set but cannot be read: %w", err)
	}
	return pulsar.NewAuthenticationTokenFromFile(p.TokenFile), nil
}
