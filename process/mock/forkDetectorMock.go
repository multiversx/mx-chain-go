package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type ForkDetectorMock struct {
	AddHeaderCalled     func(header *block.Header, hash []byte, isProcessed bool) error
	RemoveHeadersCalled func(nonce uint64)
	CheckForkCalled     func() bool
}

func (fdm *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return fdm.AddHeaderCalled(header, hash, isProcessed)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64) {
	fdm.RemoveHeadersCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() bool {
	return fdm.CheckForkCalled()
}
