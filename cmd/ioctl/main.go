package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/puzl-cloud/go-cgroup2/pkg/cgroup2io"
)

func main() {
	// Define command-line flags
	configFile := flag.String("config", "", "Path to JSON configuration file")
	cgroupPath := flag.String("cgroup", "", "Path to cgroup directory")

	flag.Parse()

	// Check required flags
	if *configFile == "" || *cgroupPath == "" {
		fmt.Println("Usage: throttleapply -config <config-file> -cgroup <cgroup-path>")
		fmt.Println("Example: throttleapply -config throttle.json -cgroup /sys/fs/cgroup/system.slice/my-service.service")
		os.Exit(1)
	}

	// Read configuration file
	jsonData, err := os.ReadFile(*configFile)
	if err != nil {
		log.Fatalf("Error reading configuration file: %v", err)
	}

	log.Printf("Applying IO limits to cgroup: %s\n", *cgroupPath)

	// Apply configuration to cgroup
	err = cgroup2io.ApplyFromJSONBytes(jsonData, *cgroupPath)
	if err != nil {
		log.Fatalf("Failed to apply IO limits: %v\n", err)
	}

	log.Println("IO limits successfully applied")
}
