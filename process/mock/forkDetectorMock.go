package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type ForkDetectorMock struct {
	AddHeaderCalled                  func(header *block.Header, hash []byte, isProcessed bool) error
	RemoveHeadersCalled              func(nonce uint64)
	CheckForkCalled                  func() (bool, uint64)
	GetHighestSignedBlockNonceCalled func() uint64
	GetHighestFinalBlockNonceCalled  func() uint64
}

func (fdm *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return fdm.AddHeaderCalled(header, hash, isProcessed)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64) {
	fdm.RemoveHeadersCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() (bool, uint64) {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestSignedBlockNonce() uint64 {
	return fdm.GetHighestSignedBlockNonceCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return fdm.GetHighestFinalBlockNonceCalled()
}
