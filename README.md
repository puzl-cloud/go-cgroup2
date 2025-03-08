# go-cgroup2

A Go library for configuring cgroup v2.

## Features

- Configure I/O throttling for block devices in cgroup v2

## Installation

```bash
go get github.com/puzl-cloud/go-cgroup2
```

## Quick Start

```go
package main

import (
	"fmt"
	"github.com/puzl-cloud/go-cgroup2/pkg/cgroup2io"
)

func main() {
	// Example JSON configuration
	jsonConfig := `{
		"Classes": {
			"example": [
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

	// Apply configuration to a cgroup
	cgroupPath := "/sys/fs/cgroup/system.slice/my-service.service"
	err := cgroup2io.ApplyFromJSONString(jsonConfig, cgroupPath)
	if err != nil {
		fmt.Printf("Failed to apply IO limits: %v\n", err)
		return
	}

	fmt.Println("Successfully applied IO limits")
}
```

## Configuration Format

The package uses JSON configuration with the following structure:

```json
{
  "Classes": {
    "example": [
      {
        "Devices": ["/dev/sda", "/dev/sdb"],
        "ThrottleReadIOPS": "1000",
        "ThrottleWriteIOPS": "800",
        "ThrottleReadBps": "50M",
        "ThrottleWriteBps": "30M"
      }
    ]
  }
}
```

### Supported Value Formats

- IOPS: Numeric values (e.g., "1000")
- Bandwidth: Values with K, M, G suffixes (e.g., "10M" for 10 megabytes per second)
- Special values: "max" or "unlimited" to set no limit

## Documentation

See the [docs](./docs/) directory for detailed documentation.
