package kamatera

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNetworkResourceSchema(t *testing.T) {
	schema := resourceNetwork().Schema["name"]
	tests := []struct {
		name            string
		input           string
		expectedSuccess bool
	}{
		{"Valid: simple name", "my-network", true},
		{"Valid: digits and dash", "net-.1234", true},
		{"Invalid: uppercase", "MyNetwork", false},
		{"Invalid: underscore", "my_network", false},
		{"Invalid: special char", "net@work", false},
		{"Invalid: too short", "", false},
		{"Invalid: too short", "a", false},
		{"Invalid: too short", "aa", false},
		{"Valid: just right", "aaa", true},
		{"Valid: just right", "aaaaaaaaaaaaaaaaaaaa", true},
		{"Invalid: too long", "aaaaaaaaaaaaaaaaaaaaa", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warns, errs := schema.ValidateFunc(tt.input, "name")
			assert.Len(t, warns, 0, "Expected no warnings for input: %s", tt.input)
			if tt.expectedSuccess {
				assert.Len(t, errs, 0, "Expected no errors for input: %s", tt.input)
			} else {
				assert.Greater(t, len(errs), 0, "Expected errors for input: %s", tt.input)
			}
		})
	}
}
