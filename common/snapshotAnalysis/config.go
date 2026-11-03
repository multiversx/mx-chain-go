package snapshotAnalysis

import "time"

// Config holds the configuration of the state snapshot analysis collector
type Config struct {
	Enabled                 bool
	Label                   string
	OutputDir               string
	SampleIntervalMs        int
	AfterSnapshotSeconds    int
	NodeHashSampleShift     int
	MaxTrackedDataTrieRoots int
}

func (c Config) withDefaults() Config {
	if c.Label == "" {
		c.Label = "unlabeled"
	}
	if c.OutputDir == "" {
		c.OutputDir = "snapshot-analysis"
	}
	if c.SampleIntervalMs <= 0 {
		c.SampleIntervalMs = 1000
	}
	if c.AfterSnapshotSeconds < 0 {
		c.AfterSnapshotSeconds = 0
	}
	if c.NodeHashSampleShift <= 0 {
		c.NodeHashSampleShift = 8
	}
	if c.NodeHashSampleShift > 16 {
		c.NodeHashSampleShift = 16
	}
	if c.MaxTrackedDataTrieRoots <= 0 {
		c.MaxTrackedDataTrieRoots = 4_000_000
	}
	return c
}

func (c Config) sampleInterval() time.Duration {
	return time.Duration(c.SampleIntervalMs) * time.Millisecond
}

func (c Config) afterWindow() time.Duration {
	return time.Duration(c.AfterSnapshotSeconds) * time.Second
}
