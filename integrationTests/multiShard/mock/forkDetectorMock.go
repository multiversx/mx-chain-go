package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type ForkDetectorMock struct {
	AddHeaderCalled             func(header *block.Header, hash []byte, isReceived bool) error
	RemoveProcessedHeaderCalled func(nonce uint64) error
	CheckForkCalled             func() bool
}

func (fdm *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isReceived bool) error {
	return fdm.AddHeaderCalled(header, hash, isReceived)
}

func (fdm *ForkDetectorMock) RemoveProcessedHeader(nonce uint64) error {
	return fdm.RemoveProcessedHeaderCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() bool {
	return fdm.CheckForkCalled()
}
