package cgroup2io

import (
	"testing"
)

func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		hasError bool
	}{
		{"EmptyValue", "", 0, true},
		{"NumericValue", "1000", 1000, false},
		{"KiloValue", "10k", 10 * 1000, false},
		{"MegaValue", "5M", 5 * 1000 * 1000, false},
		{"GigaValue", "1G", 1000 * 1000 * 1000, false},
		{"MaxValue", "max", MaxInt64Value, false},
		{"UnlimitedValue", "unlimited", MaxInt64Value, false},
		{"InvalidValue", "10X", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseValue(tc.input)

			if tc.hasError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tc.hasError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tc.hasError && result != tc.expected {
				t.Errorf("Expected %d but got %d", tc.expected, result)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	validJSON := []byte(`{
        "Classes": {
            "exmaple": [
                {
                    "Devices": ["/dev/null"],
                    "ThrottleReadIOPS": "1000",
                    "ThrottleWriteIOPS": "800",
                    "ThrottleReadBps": "50M",
                    "ThrottleWriteBps": "30M"
                }
            ]
        }
    }`)

	// Testing custom class name
	customClassJSON := []byte(`{
        "Classes": {
            "custom_policy": [
                {
                    "Devices": ["/dev/null"],
                    "ThrottleReadIOPS": "1000",
                    "ThrottleWriteIOPS": "800"
                }
            ]
        }
    }`)

	invalidJSON := []byte(`{
        "Classes": {}
    }`)

	emptyClassJSON := []byte(`{
        "Classes": {
            "exmaple": []
        }
    }`)

	missingDevicesJSON := []byte(`{
        "Classes": {
            "exmaple": [
                {
                    "ThrottleReadIOPS": "1000"
                }
            ]
        }
    }`)

	// Test valid JSON
	config, err := ParseConfig(validJSON)
	if err != nil {
		t.Errorf("Failed to parse valid JSON: %v", err)
	}

	quotaConfigs, ok := config.Classes["exmaple"]
	if !ok {
		t.Errorf("Expected 'quota' class to exist")
	}
	if len(quotaConfigs) != 1 {
		t.Errorf("Expected 1 quota config, got %d", len(quotaConfigs))
	}

	// Test custom class name
	customConfig, err := ParseConfig(customClassJSON)
	if err != nil {
		t.Errorf("Failed to parse custom class JSON: %v", err)
	}

	customConfigs, ok := customConfig.Classes["custom_policy"]
	if !ok {
		t.Errorf("Expected 'custom_policy' class to exist")
	}
	if len(customConfigs) != 1 {
		t.Errorf("Expected 1 custom policy config, got %d", len(customConfigs))
	}

	// Test defaults
	defaultJSON := []byte(`{
        "Classes": {
            "exmaple": [
                {
                    "Devices": ["/dev/null"]
                }
            ]
        }
    }`)

	defaultConfig, err := ParseConfig(defaultJSON)
	if err != nil {
		t.Errorf("Failed to parse default JSON: %v", err)
	}

	quotaConfigs = defaultConfig.Classes["exmaple"]
	quota := quotaConfigs[0]
	if quota.ThrottleReadIOPS != "max" {
		t.Errorf("Expected default ReadIOPS %s, got %s", "max", quota.ThrottleReadIOPS)
	}

	if quota.ThrottleWriteIOPS != "max" {
		t.Errorf("Expected default WriteIOPS %s, got %s", "max", quota.ThrottleWriteIOPS)
	}

	if quota.ThrottleReadBps != "max" {
		t.Errorf("Expected default ReadBps %s, got %s", "max", quota.ThrottleReadBps)
	}

	if quota.ThrottleWriteBps != "max" {
		t.Errorf("Expected default WriteBps %s, got %s", "max", quota.ThrottleWriteBps)
	}

	// Test completely invalid JSON
	_, err = ParseConfig(invalidJSON)
	if err == nil {
		t.Errorf("Expected error for empty Classes map, got nil")
	}

	// Test empty class array
	_, err = ParseConfig(emptyClassJSON)
	if err == nil {
		t.Errorf("Expected error for empty quota array, got nil")
	}

	// Test missing devices
	_, err = ParseConfig(missingDevicesJSON)
	if err == nil {
		t.Errorf("Expected error for missing devices, got nil")
	}
}

// TestFormatMaxOrValue tests the formatMaxOrValue function
func TestFormatMaxOrValue(t *testing.T) {
	// Test max value
	result := formatMaxOrValue(MaxInt64Value)
	expected := "max"
	if result != expected {
		t.Errorf("Expected %s for MaxInt64Value, got %s", expected, result)
	}

	// Test normal value
	result = formatMaxOrValue(1000)
	expected = "1000"
	if result != expected {
		t.Errorf("Expected %s for 1000, got %s", expected, result)
	}
}
