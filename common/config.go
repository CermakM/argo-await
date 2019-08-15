package common

import (
	"github.com/sirupsen/logrus"
	"os"
)

// ENV VARS
const (
	// EnvVarKubeConfig is the path to the Kubernetes configuration
	EnvVarKubeConfig = "KUBE_CONFIG"

	// EnvVarDebugLog is the env var to turn on the debug mode for logging
	EnvVarDebugLog = "DEBUG_LOG"
)

// Logger initializes a new logrus.Logger
func Logger() *logrus.Logger {

	log := &logrus.Logger{
		Out:   os.Stderr,
		Level: logrus.InfoLevel,
		Formatter: &logrus.TextFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			FullTimestamp:    true,
			ForceColors:      true,
			QuoteEmptyFields: true,
		},
	}

	debugMode, ok := os.LookupEnv(EnvVarDebugLog)
	if ok && debugMode == "true" {
		log.SetLevel(logrus.DebugLevel)
	}

	return log
}
