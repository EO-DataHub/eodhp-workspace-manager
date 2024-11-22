package utils

import (
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// InitLogger configures and initializes the logger based on the provided log level
func InitLogger(level string) {
	// Use UTC timestamps for logging
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	// Set the global log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

}
