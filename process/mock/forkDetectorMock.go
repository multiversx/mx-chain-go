package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type ForkDetectorMock struct {
	AddHeaderCalled                 func(header data.HeaderHandler, hash []byte, isProcessed bool) error
	RemoveHeadersCalled             func(nonce uint64)
	CheckForkCalled                 func() (bool, uint64)
	GetHighestFinalBlockNonceCalled func() uint64
	ProbableHighestNonceCalled      func() uint64
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, isProcessed bool) error {
	return fdm.AddHeaderCalled(header, hash, isProcessed)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64) {
	fdm.RemoveHeadersCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() (bool, uint64) {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return fdm.GetHighestFinalBlockNonceCalled()
}

func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}
