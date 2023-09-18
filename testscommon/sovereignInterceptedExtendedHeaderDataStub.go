package testscommon

import "github.com/multiversx/mx-chain-core-go/data"

type SovereignInterceptedExtendedHeaderDataStub struct {
	*InterceptedDataStub
	HashCalled              func() []byte
	GetExtendedHeaderCalled func() data.ShardHeaderExtendedHandler
}

func (s *SovereignInterceptedExtendedHeaderDataStub) Hash() []byte {
	if s.HashCalled != nil {
		return s.HashCalled()
	}

	return nil
}

func (s *SovereignInterceptedExtendedHeaderDataStub) GetExtendedHeader() data.ShardHeaderExtendedHandler {
	if s.GetExtendedHeaderCalled != nil {
		return s.GetExtendedHeaderCalled()
	}

	return nil
}
