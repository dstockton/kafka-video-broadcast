package utils

import (
	"os"
	"testing"
)

func TestInitialiseLogging(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
	}{
		{
			name: "InitialiseLogging - Default",
		},
		{
			name:     "InitialiseLogging - DEBUG",
			logLevel: "DEBUG",
		},
		{
			name:     "InitialiseLogging - Invalid",
			logLevel: "Invalid",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if test.logLevel != "" {
				os.Setenv("LOG_LEVEL", test.logLevel)
			}
			InitialiseLogging()
		})
	}
}
