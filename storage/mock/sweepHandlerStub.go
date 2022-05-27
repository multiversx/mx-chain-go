package mock

// EvictionHandlerStub -
type EvictionHandlerStub struct {
	EvictedCalled func(key []byte)
}

// Evicted -
func (sh *EvictionHandlerStub) Evicted(key []byte) {
	if sh.EvictedCalled != nil {
		sh.EvictedCalled(key)
	}
}
