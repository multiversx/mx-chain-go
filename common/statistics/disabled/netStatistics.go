package disabled

const notAvailable = "N/A"

type netStatistics struct {
}

// NewDisabledNetStatistics creates a new disabled network statistics
func NewDisabledNetStatistics() *netStatistics {
	return &netStatistics{}
}

// TotalSentInCurrentEpoch will send the notAvailable constant string
func (stats *netStatistics) TotalSentInCurrentEpoch() string {
	return notAvailable
}

// TotalReceivedInCurrentEpoch will send the notAvailable constant string
func (stats *netStatistics) TotalReceivedInCurrentEpoch() string {
	return notAvailable
}

// EpochConfirmed does nothing
func (stats *netStatistics) EpochConfirmed(_ uint32, _ uint64) {
}

// Close returns nil
func (stats *netStatistics) Close() error {
	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (stats *netStatistics) IsInterfaceNil() bool {
	return stats == nil
}
