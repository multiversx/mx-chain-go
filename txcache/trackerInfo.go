package txcache

type trackerDiagnosis struct {
	numTrackedBlocks   uint64
	numTrackedAccounts uint64
}

// NewTrackerDiagnosis creates a new trackerDiagnosis
func NewTrackerDiagnosis(numTrackedBlocks uint64, numTrackedAccounts uint64) *trackerDiagnosis {
	return &trackerDiagnosis{
		numTrackedBlocks:   numTrackedBlocks,
		numTrackedAccounts: numTrackedAccounts,
	}
}

// GetNumTrackedBlocks returns the number of tracked blocks
func (t *trackerDiagnosis) GetNumTrackedBlocks() uint64 {
	return t.numTrackedBlocks
}

// GetNumTrackedAccounts returns the number of tracked accounts
func (t *trackerDiagnosis) GetNumTrackedAccounts() uint64 {
	return t.numTrackedAccounts
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trackerDiagnosis) IsInterfaceNil() bool {
	return t == nil
}
