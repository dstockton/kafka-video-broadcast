package utils

import (
	log "github.com/sirupsen/logrus"
)

// InitialiseLogging initializes logging
func InitialiseLogging() {
	lvl := GetEnv("LOG_LEVEL", "INFO")

	ll, err := log.ParseLevel(lvl)
	if err != nil {
		ll = log.InfoLevel
	}

	// set global log level
	log.SetLevel(ll)

	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05.000", FullTimestamp: true})
}
