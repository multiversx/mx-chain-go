package disabled

type trieSyncStatisticsProvider struct {
}

// NewTrieSyncStatisticsProvider returns a new instance of trieSyncStatisticsProvider
func NewTrieSyncStatisticsProvider() *trieSyncStatisticsProvider {
	return &trieSyncStatisticsProvider{}
}

// NumReceived returns 0 always
func (tssp *trieSyncStatisticsProvider) NumReceived() int {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (tssp *trieSyncStatisticsProvider) IsInterfaceNil() bool {
	return tssp == nil
}
