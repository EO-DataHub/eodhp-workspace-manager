package utils

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

type PulsarConfig struct {
	URL           string `yaml:"url"`
	TopicProducer string `yaml:"topicProducer"`
	TopicConsumer string `yaml:"topicConsumer"`
	Subscription  string `yaml:"subscription"`
}

type AWSConfig struct {
	Cluster string `yaml:"cluster"`
	FSID    string `yaml:"fsId"`
}

type StorageConfig struct {
	Size         string `yaml:"size"`
	StorageClass string `yaml:"storageClass"`
	PVCName      string `yaml:"pvcName"`
	Driver       string `yaml:"efs.csi.aws.com"`
}

// Config holds the application's configuration
type Config struct {
	LogLevel string        `yaml:"logLevel"`
	Pulsar   PulsarConfig  `yaml:"pulsar"`
	AWS      AWSConfig     `yaml:"aws"`
	Storage  StorageConfig `yaml:"storage"`
}

// LoadConfig loads the application configuration from a file
func LoadConfig() *Config {
	// Get the config file path from the environment variable
	configPath := "/home/jlangstone/Work/UKEODHP/eodhp-workspace-manager/tmp/config.yaml"

	// Parse the configuration file as a template
	tmpl, err := template.ParseFiles(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing configuration file template")
	}

	// Load environment variables into a map
	envVars := loadEnvVars()

	// Execute the template with the environment variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, envVars); err != nil {
		log.Fatal().Err(err).Msg("Error executing configuration file template")
	}

	// Load the config from the processed template
	config := &Config{}
	if err := yaml.Unmarshal(buf.Bytes(), config); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal configuration file")
	}

	return config
}

// loadEnvVars loads environment variables into a map
func loadEnvVars() map[string]string {
	envVars := make(map[string]string)
	for _, env := range os.Environ() {
		kv := strings.SplitN(env, "=", 2)
		if len(kv) == 2 {
			envVars[kv[0]] = kv[1]
		}
	}
	return envVars
}
