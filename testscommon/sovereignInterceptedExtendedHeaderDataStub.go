package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

// SovereignInterceptedExtendedHeaderDataStub -
type SovereignInterceptedExtendedHeaderDataStub struct {
	*InterceptedDataStub
	HashCalled              func() []byte
	GetExtendedHeaderCalled func() data.ShardHeaderExtendedHandler
}

// Hash -
func (s *SovereignInterceptedExtendedHeaderDataStub) Hash() []byte {
	if s.HashCalled != nil {
		return s.HashCalled()
	}

	return nil
}

// GetExtendedHeader -
func (s *SovereignInterceptedExtendedHeaderDataStub) GetExtendedHeader() data.ShardHeaderExtendedHandler {
	if s.GetExtendedHeaderCalled != nil {
		return s.GetExtendedHeaderCalled()
	}

	return nil
}
