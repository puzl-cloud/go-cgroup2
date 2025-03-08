package main

import (
	"fmt"
	"log"
	"os"

	"github.com/puzl-cloud/go-cgroup2/pkg/cgroup2io"
)

func main() {
	// Check if the cgroup path is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <cgroup-path>")
		fmt.Println("Example: go run main.go /sys/fs/cgroup/system.slice/my-service.service")
		os.Exit(1)
	}

	cgroupPath := os.Args[1]

	// Example JSON configuration
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

	log.Printf("Applying IO limits to cgroup: %s\n", cgroupPath)

	// Apply configuration to the cgroup
	err := cgroup2io.ApplyFromJSONString(jsonConfig, cgroupPath)
	if err != nil {
		log.Fatalf("Failed to apply IO limits: %v\n", err)
	}

	log.Println("Successfully applied IO limits")
}
