package mock

import "github.com/ElrondNetwork/elrond-go/process"

// StatusCoreComponentsMock -
type StatusCoreComponentsMock struct {
	TrieSyncStatisticsField process.TrieSyncStatisticsProvider
}

// TrieSyncStatistics -
func (sccm *StatusCoreComponentsMock) TrieSyncStatistics() process.TrieSyncStatisticsProvider {
	return sccm.TrieSyncStatisticsField
}

// IsInterfaceNil -
func (sccm *StatusCoreComponentsMock) IsInterfaceNil() bool {
	return sccm == nil
}
