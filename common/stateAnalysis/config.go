package stateAnalysis

// Config holds the configuration of the state analysis collector
type Config struct {
	Enabled                  bool
	Label                    string
	OutputDir                string
	BlocksPerRow             int
	MaxTrackedKeys           int
	TopKPerEpoch             int
	MemorySampleEveryNBlocks int
	OutlierCommitTimeMs      int64
	OutlierStorageReads      int64
	OutlierJournalLen        int64
	ProfileEveryNEpochs      int
	ProfileNumBlocks         int
}

func (c Config) withDefaults() Config {
	if c.OutputDir == "" {
		c.OutputDir = "state-analysis"
	}
	if c.Label == "" {
		c.Label = "unlabeled"
	}
	if c.BlocksPerRow <= 0 {
		c.BlocksPerRow = 100
	}
	if c.MaxTrackedKeys <= 0 {
		c.MaxTrackedKeys = 2_000_000
	}
	if c.TopKPerEpoch <= 0 {
		c.TopKPerEpoch = 1000
	}
	if c.MemorySampleEveryNBlocks <= 0 {
		c.MemorySampleEveryNBlocks = 100
	}
	if c.OutlierCommitTimeMs <= 0 {
		c.OutlierCommitTimeMs = 2000
	}
	if c.OutlierStorageReads <= 0 {
		c.OutlierStorageReads = 50000
	}
	if c.OutlierJournalLen <= 0 {
		c.OutlierJournalLen = 50000
	}
	return c
}
