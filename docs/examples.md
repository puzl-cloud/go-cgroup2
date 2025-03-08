# Usage Examples

This document provides detailed examples of how to use the `cgroup2io` package for various use cases.

## Basic Usage

The most common use case is to apply IO limits to a specific cgroup:

```go
package main

import (
	"fmt"
	"github.com/puzl-cloud/go-cgroup2/pkg/cgroup2io"
)

func main() {
	// JSON configuration
	jsonConfig := `{
		"Classes": {
			"quota": [
				{
					"Devices": ["/dev/sda"],
					"ThrottleReadIOPS": "1000",
					"ThrottleWriteIOPS": "800",
					"ThrottleReadBps": "50M",
					"ThrottleWriteBps": "30M"
				}
			]
		}
	}`

	// Apply configuration
	cgroupPath := "/sys/fs/cgroup/system.slice/my-service.service"
	err := cgroup2io.ApplyFromJSONString(jsonConfig, cgroupPath)
	if err != nil {
		fmt.Printf("Failed to apply IO limits: %v\n", err)
		return
	}

	fmt.Println("Successfully applied IO limits")
}
```

## Using Special Values

You can use special values like "max" or "unlimited" to remove restrictions:

```go
jsonConfig := `{
	"Classes": {
		"quota": [
			{
				"Devices": ["/dev/sda"],
				"ThrottleReadIOPS": "max",
				"ThrottleWriteIOPS": "unlimited",
				"ThrottleReadBps": "50M",
				"ThrottleWriteBps": "30M"
			}
		]
	}
}`
```

## Multiple Devices Configuration

```go
jsonConfig := `{
	"Classes": {
		"quota": [
			{
				"Devices": ["/dev/sda", "/dev/sdb"],
				"ThrottleReadIOPS": "1000",
				"ThrottleWriteIOPS": "800",
				"ThrottleReadBps": "50M",
				"ThrottleWriteBps": "30M"
			},
			{
				"Devices": ["/dev/sdc"],
				"ThrottleReadIOPS": "500",
				"ThrottleWriteIOPS": "300",
				"ThrottleReadBps": "20M",
				"ThrottleWriteBps": "10M"
			}
		]
	}
}`
```

## Programmatic Configuration

Instead of JSON strings, you can build the configuration programmatically:

```go
config := &cgroup2io.Config{}
config.Classes.Quota = []cgroup2io.QuotaConfig{
	{
		Devices:           []string{"/dev/sda"},
		ThrottleReadIOPS:  "1000",
		ThrottleWriteIOPS: "800",
		ThrottleReadBps:   "50M",
		ThrottleWriteBps:  "30M",
	},
}

err := cgroup2io.ApplyIOLimits(cgroupPath, config)
if err != nil {
	fmt.Printf("Failed to apply IO limits: %v\n", err)
	return
}
```

## Integration with Kubernetes

This example shows how to use the package with Kubernetes to apply limits based on resource requests:

```go
package main

import (
	"fmt"
	"strconv"
	
	"github.com/puzl-cloud/go-cgroup2/pkg/cgroup2io"
	"k8s.io/apimachinery/pkg/api/resource"
)

func applyPodIOLimits(podUID, containerName string, cpuLimit, memLimit resource.Quantity) error {
	// Calculate IO limits based on resource allocations
	// This is just an example - actual calculation would depend on your specific requirements
	cpuMilliValue := cpuLimit.MilliValue()
	memBytes := memLimit.Value()
	
	// Simple formula: 100 IOPS per CPU core, 10MB/s per GB of memory
	readIOPS := strconv.FormatInt(cpuMilliValue/10, 10) // 100 IOPS per CPU core
	writeIOPS := strconv.FormatInt(cpuMilliValue/20, 10) // 50 IOPS per CPU core
	
	// Calculate bandwidth: 10MB/s per GB of memory
	memGB := float64(memBytes) / (1024 * 1024 * 1024)
	readBW := fmt.Sprintf("%.0fM", memGB*10)
	writeBW := fmt.Sprintf("%.0fM", memGB*5)
	
	// Create config
	config := &cgroup2io.Config{}
	config.Classes.Quota = []cgroup2io.QuotaConfig{
		{
			Devices:           []string{"/dev/sda"},
			ThrottleReadIOPS:  readIOPS,
			ThrottleWriteIOPS: writeIOPS,
			ThrottleReadBps:   readBW,
			ThrottleWriteBps:  writeBW,
		},
	}
	
	// Path to container's cgroup
	cgroupPath := fmt.Sprintf("/sys/fs/cgroup/kubepods/pod%s/%s", podUID, containerName)
	
	return cgroup2io.ApplyIOLimits(cgroupPath, config)
}
```