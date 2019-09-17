package poolscleaner

// NilPoolsCleaner will be used when an PoolsCleaner is required, but another one isn't necessary or available
type NilPoolsCleaner struct {
}

// NewNilPoolsCleaner will return an instance of the struct
func NewNilPoolsCleaner() *NilPoolsCleaner {
	return new(NilPoolsCleaner)
}

// Clean method - won't do anything
func (nsh *NilPoolsCleaner) Clean(haveTime func() bool) error {
	return nil
}

// NumRemovedTxs - won't do anything
func (nsh *NilPoolsCleaner) NumRemovedTxs() uint64 {
	return 0
}

// IsInterfaceNil returns true if there is no value under the interface
func (nsh *NilPoolsCleaner) IsInterfaceNil() bool {
	if nsh == nil {
		return true
	}
	return false
}
