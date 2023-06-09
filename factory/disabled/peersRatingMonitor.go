package disabled

type peersRatingMonitor struct {
}

// NewPeersRatingMonitor returns a new disabled PeersRatingMonitor implementation
func NewPeersRatingMonitor() *peersRatingMonitor {
	return &peersRatingMonitor{}
}

// GetConnectedPeersRatings returns an empty string as it is disabled
func (monitor *peersRatingMonitor) GetConnectedPeersRatings() string {
	return ""
}

// IsInterfaceNil returns true if there is no value under the interface
func (monitor *peersRatingMonitor) IsInterfaceNil() bool {
	return monitor == nil
}
