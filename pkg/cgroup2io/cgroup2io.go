// Package cgroupiothrottle provides functions for configuring IO limits in cgroup v2
package cgroup2io

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Default values for unspecified parameters
const (
	MaxInt64Value = math.MaxInt64 // Represents "max" value for cgroup
)

// Config represents the JSON configuration structure with flexible class names
type Config struct {
	Classes map[string][]QuotaConfig `json:"Classes"`
}

// QuotaConfig defines throttling parameters
type QuotaConfig struct {
	Devices           []string `json:"Devices"`
	ThrottleReadIOPS  string   `json:"ThrottleReadIOPS"`
	ThrottleWriteIOPS string   `json:"ThrottleWriteIOPS"`
	ThrottleReadBps   string   `json:"ThrottleReadBps"`
	ThrottleWriteBps  string   `json:"ThrottleWriteBps"`
}

// ParseConfig parses JSON data into Config structure
func ParseConfig(jsonData []byte) (*Config, error) {
	var config Config
	err := json.Unmarshal(jsonData, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %w", err)
	}

	// Validate config
	if len(config.Classes) == 0 {
		return nil, fmt.Errorf("no class configurations found in JSON")
	}

	// Apply defaults to missing values for all classes
	for className, quotaConfigs := range config.Classes {
		if len(quotaConfigs) == 0 {
			return nil, fmt.Errorf("no quota configurations found in class %s", className)
		}

		for i := range quotaConfigs {
			if len(quotaConfigs[i].Devices) == 0 {
				return nil, fmt.Errorf("devices list cannot be empty in %s[%d]", className, i)
			}

			// Apply defaults if values are empty
			if quotaConfigs[i].ThrottleReadIOPS == "" {
				quotaConfigs[i].ThrottleReadIOPS = "max"
			}
			if quotaConfigs[i].ThrottleWriteIOPS == "" {
				quotaConfigs[i].ThrottleWriteIOPS = "max"
			}

			if quotaConfigs[i].ThrottleReadBps == "" {
				quotaConfigs[i].ThrottleReadBps = "max"
			}

			if quotaConfigs[i].ThrottleWriteBps == "" {
				quotaConfigs[i].ThrottleWriteBps = "max"
			}
		}
	}

	return &config, nil
}

// ParseConfigFromString parses JSON string into Config structure
func ParseConfigFromString(jsonStr string) (*Config, error) {
	return ParseConfig([]byte(jsonStr))
}

// ParseValue converts values with suffixes (K, M, G) to numbers,
// using the resource package from Kubernetes.
// Special values "max" or "unlimited" return MaxInt64Value.
func ParseValue(value string) (int64, error) {
	// Handle special values
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("empty value not allowed")
	}

	// Check for unlimited/max values
	if strings.EqualFold(value, "max") || strings.EqualFold(value, "unlimited") {
		return MaxInt64Value, nil
	}

	// For values with units (K, M, G), use k8s resource
	quantity, err := resource.ParseQuantity(value)
	if err != nil {
		return 0, fmt.Errorf("failed to parse value %s: %w", value, err)
	}

	return quantity.Value(), nil
}

// GetDeviceNumbers retrieves major:minor from device path
func GetDeviceNumbers(devicePath string) (string, error) {
	stat, err := os.Stat(devicePath)
	if err != nil {
		return "", fmt.Errorf("failed to get device info for %s: %w", devicePath, err)
	}

	statT := stat.Sys().(*syscall.Stat_t)
	major := statT.Rdev >> 8
	minor := statT.Rdev & 0xFF

	return fmt.Sprintf("%d:%d", major, minor), nil
}

// ApplyIOLimits applies IO limits for cgroup v2 based on configuration
func ApplyIOLimits(cgroupPath string, config *Config) error {
	// Check if cgroup directory exists
	if _, err := os.Stat(cgroupPath); os.IsNotExist(err) {
		return fmt.Errorf("cgroup path does not exist: %s", cgroupPath)
	}

	for className, quotaConfigs := range config.Classes {
		for i, quota := range quotaConfigs {
			// Process each device
			for _, device := range quota.Devices {
				// Get major:minor for the device
				devNumbers, err := GetDeviceNumbers(device)
				if err != nil {
					return fmt.Errorf("failed to get device numbers for %s: %w", device, err)
				}

				// Parse limit values
				readBps, err := ParseValue(quota.ThrottleReadBps)
				if err != nil {
					return fmt.Errorf("error parsing ThrottleReadBps in %s[%d]: %w", className, i, err)
				}

				writeBps, err := ParseValue(quota.ThrottleWriteBps)
				if err != nil {
					return fmt.Errorf("error parsing ThrottleWriteBps in %s[%d]: %w", className, i, err)
				}

				readIOPS, err := ParseValue(quota.ThrottleReadIOPS)
				if err != nil {
					return fmt.Errorf("error parsing ThrottleReadIOPS in %s[%d]: %w", className, i, err)
				}

				writeIOPS, err := ParseValue(quota.ThrottleWriteIOPS)
				if err != nil {
					return fmt.Errorf("error parsing ThrottleWriteIOPS in %s[%d]: %w", className, i, err)
				}

				// Apply limits to the device
				err = ApplyDeviceIOLimits(cgroupPath, devNumbers, readBps, writeBps, readIOPS, writeIOPS)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ApplyDeviceIOLimits applies IO limits for a specific device
func ApplyDeviceIOLimits(cgroupPath, deviceNumbers string, readBps, writeBps, readIOPS, writeIOPS int64) error {
	ioMaxPath := filepath.Join(cgroupPath, "io.max")

	// Read current IO limits
	content, err := os.ReadFile(ioMaxPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", ioMaxPath, err)
	}

	lines := strings.Split(string(content), "\n")
	newLines := []string{}
	deviceFound := false

	// Format new limit values for the device
	rbpsStr := formatMaxOrValue(readBps)
	wbpsStr := formatMaxOrValue(writeBps)
	riopsStr := formatMaxOrValue(readIOPS)
	wiopsStr := formatMaxOrValue(writeIOPS)

	newLimit := fmt.Sprintf("%s rbps=%s wbps=%s riops=%s wiops=%s",
		deviceNumbers, rbpsStr, wbpsStr, riopsStr, wiopsStr)

	// Update existing limits if the device is already configured
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) > 0 && parts[0] == deviceNumbers {
			newLines = append(newLines, newLimit)
			deviceFound = true
		} else {
			newLines = append(newLines, line)
		}
	}

	// Add new device configuration if not found in existing settings
	if !deviceFound {
		newLines = append(newLines, newLimit)
	}

	// Write back all configurations to io.max
	// Note: No need to specify permissions as the file already exists
	return os.WriteFile(ioMaxPath, []byte(strings.Join(newLines, "\n")), 0)
}

// formatMaxOrValue converts MaxInt64Value to "max" string or returns the numeric value
func formatMaxOrValue(value int64) string {
	if value == MaxInt64Value {
		return "max"
	}
	return fmt.Sprintf("%d", value)
}

// ApplyFromJSONString applies IO limits from a JSON string
func ApplyFromJSONString(jsonStr string, cgroupPath string) error {
	config, err := ParseConfigFromString(jsonStr)
	if err != nil {
		return err
	}

	return ApplyIOLimits(cgroupPath, config)
}

// ApplyFromJSONBytes applies IO limits from JSON bytes
func ApplyFromJSONBytes(jsonData []byte, cgroupPath string) error {
	config, err := ParseConfig(jsonData)
	if err != nil {
		return err
	}

	return ApplyIOLimits(cgroupPath, config)
}

// ValidateConfig checks if the config structure is valid
func ValidateConfig(config *Config) error {
	if len(config.Classes) == 0 {
		return fmt.Errorf("no class configurations found")
	}

	for className, quotaConfigs := range config.Classes {
		if len(quotaConfigs) == 0 {
			return fmt.Errorf("no quota configurations found in class %s", className)
		}

		for i, quota := range quotaConfigs {
			if len(quota.Devices) == 0 {
				return fmt.Errorf("devices list cannot be empty in %s[%d]", className, i)
			}

			// Validate that devices exist
			for _, dev := range quota.Devices {
				if _, err := os.Stat(dev); os.IsNotExist(err) {
					return fmt.Errorf("device %s does not exist", dev)
				}
			}
		}
	}

	return nil
}
