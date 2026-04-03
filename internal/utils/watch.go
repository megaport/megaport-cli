package utils

import "time"

// WatchConfig holds configuration for the watch loop.
type WatchConfig struct {
	Interval     time.Duration
	NoColor      bool
	OutputFormat string
	ResourceType string // "Port", "VXC", "MCR", "MVE"
	ResourceUID  string
}
