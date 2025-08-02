package kamatera

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestResourceServerSchemaValidation(t *testing.T) {
	for _, tt := range []struct {
		attr                 string
		value                interface{}
		expectedErrorStrings []string
	}{
		// name
		{"name", "", []string{
			"expected length of name to be in the range (4 - 40), got ",
		}},
		{"name", "aaa", []string{
			"expected length of name to be in the range (4 - 40), got aaa",
		}},
		{"name", "aaaa", []string{}},
		{"name", strings.Repeat("a", 40), []string{}},
		{"name", strings.Repeat("a", 41), []string{
			"expected length of name to be in the range (4 - 40), got aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		}},
		{"name", "@" + strings.Repeat("a", 41), []string{
			"expected length of name to be in the range (4 - 40), got @aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"invalid value for name (must contain only letters, digits, dashes (-) and dots (.))",
		}},
		{"name", "123abc@", []string{
			"invalid value for name (must contain only letters, digits, dashes (-) and dots (.))",
		}},
		{"name", "123abc", []string{}},

		// datacenter id
		{"datacenter_id", "a", []string{
			"expected length of datacenter_id to be in the range (2 - 6), got a",
			"invalid value for datacenter_id (must contain only uppercase letters, digits and dashes (-))",
		}},
		{"datacenter_id", "EU", []string{}},
		{"datacenter_id", "US-NY", []string{}},
		{"datacenter_id", "US-NY2", []string{}},

		// cpu type
		{"cpu_type", "C", []string{"expected cpu_type to be one of [A B T D], got C"}},
		{"cpu_type", "A", []string{}},

		// cpu cores
		{"cpu_cores", -5, []string{"expected cpu_cores to be at least (1), got -5"}},
		{"cpu_cores", 1, []string{}},

		// ram mb
		{"ram_mb", 255, []string{"expected ram_mb to be at least (256), got 255"}},
		{"ram_mb", 1024, []string{}},

		// billing cycle
		{"billing_cycle", "invalid", []string{"expected billing_cycle to be one of [hourly monthly], got invalid"}},
		{"billing_cycle", "hourly", []string{}},
		{"billing_cycle", "monthly", []string{}},
	} {
		var testName string
		if len(tt.expectedErrorStrings) == 0 {
			testName = fmt.Sprintf("valid server %s", tt.attr)
		} else {
			testName = fmt.Sprintf("invalid server %s", tt.attr)
		}
		t.Run(testName, func(t *testing.T) {
			schema := resourceServer().Schema[tt.attr]
			warns, errs := schema.ValidateFunc(tt.value, tt.attr)
			assert.Len(t, warns, 0, "Expected no warnings for attr %s value %s", tt.attr, tt.value)
			errorStrings := []string{}
			for _, err := range errs {
				errorStrings = append(errorStrings, err.Error())
			}
			assert.Equal(t, tt.expectedErrorStrings, errorStrings)
		})
	}
}
