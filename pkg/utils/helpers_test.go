package utils

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		defaultValue  string
		expectedValue string
	}{
		{
			name:          "no defaulting",
			key:           "DOES_EXIST",
			defaultValue:  "not-used",
			expectedValue: "do not default",
		},
		{
			name:          "defaulted",
			key:           "DOES_NOT_EXIST",
			defaultValue:  "this is the real deal",
			expectedValue: "this is the real deal",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			os.Setenv("DOES_EXIST", "do not default")
			gotValue := GetEnv(test.key, test.defaultValue)
			if gotValue != test.expectedValue {
				t.Errorf("Expected '%s', got '%s'", test.expectedValue, gotValue)
			}
		})
	}
}
