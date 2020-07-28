package libp2p

import "time"

// LocalSyncTimer uses the local system to provide the current time
type LocalSyncTimer struct {
}

// CurrentTime returns the local current time
func (lst *LocalSyncTimer) CurrentTime() time.Time {
	return time.Now()
}

// IsInterfaceNil returns true if there is no value under the interface
func (lst *LocalSyncTimer) IsInterfaceNil() bool {
	return lst == nil
}
