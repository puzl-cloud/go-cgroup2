# Configuration Reference

The `cgroup2io` package uses a JSON configuration format to specify I/O limits for devices.

## Configuration Structure

The JSON configuration has the following structure:

```json
{
  "Classes": {
    "example": [
      {
        "Devices": ["<device-path-1>", "<device-path-2>"],
        "ThrottleReadIOPS": "<read-iops-limit>",
        "ThrottleWriteIOPS": "<write-iops-limit>",
        "ThrottleReadBps": "<read-bandwidth-limit>",
        "ThrottleWriteBps": "<write-bandwidth-limit>"
      }
    ]
  }
}
```

## Configuration Fields

### Devices

An array of block device paths (e.g., `/dev/sda`, `/dev/sdb`) to apply limits to. The devices must exist on the system.

### ThrottleReadIOPS

Read I/O operations per second limit. Accepts:
- Numeric values (e.g., "1000")
- "max" or "unlimited" to set no limit
- Default: "max" if not specified

### ThrottleWriteIOPS

Write I/O operations per second limit. Accepts:
- Numeric values (e.g., "1000")
- "max" or "unlimited" to set no limit
- Default: "max" if not specified

### ThrottleReadBps

Read bandwidth limit in bytes per second. Accepts:
- Numeric values (e.g., "10485760" for 10MB/s)
- Values with suffixes (e.g., "10M" for 10MB/s)
- Supported suffixes: K, M, G (case-insensitive)
- "max" or "unlimited" to set no limit
- Default: "max" if not specified

### ThrottleWriteBps

Write bandwidth limit in bytes per second. Accepts:
- Numeric values (e.g., "10485760" for 10MB/s)
- Values with suffixes (e.g., "10M" for 10MB/s)
- Supported suffixes: K, M, G (case-insensitive)
- "max" or "unlimited" to set no limit
- Default: "max" if not specified

## Multiple Configurations

You can specify multiple configurations in the `example` classes to apply different limits to different devices:

```json
{
  "Classes": {
    "example": [
      {
        "Devices": ["/dev/sda"],
        "ThrottleReadIOPS": "1000",
        "ThrottleWriteIOPS": "800",
        "ThrottleReadBps": "50M",
        "ThrottleWriteBps": "30M"
      },
      {
        "Devices": ["/dev/sdb", "/dev/sdc"],
        "ThrottleReadIOPS": "500",
        "ThrottleWriteIOPS": "300",
        "ThrottleReadBps": "20M",
        "ThrottleWriteBps": "10M"
      }
    ]
  }
}
```
