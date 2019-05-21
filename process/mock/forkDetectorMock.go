package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
)

type ForkDetectorMock struct {
	AddHeaderCalled                 func(header data.HeaderHandler, hash []byte, isCommitted bool) error
	RemoveHeadersCalled             func(nonce uint64, hash []byte)
	CheckForkCalled                 func() (bool, uint64)
	GetHighestFinalBlockNonceCalled func() uint64
	ProbableHighestNonceCalled      func() uint64
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, isCommitted bool) error {
	return fdm.AddHeaderCalled(header, hash, isCommitted)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64, hash []byte) {
	fdm.RemoveHeadersCalled(nonce, hash)
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
