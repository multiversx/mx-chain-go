package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type ForkDetectorMock struct {
	AddHeaderCalled                         func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error
	RemoveHeadersCalled                     func(nonce uint64, hash []byte)
	CheckForkCalled                         func() (bool, uint64, []byte)
	GetHighestFinalBlockNonceCalled         func() uint64
	ProbableHighestNonceCalled              func() uint64
	ResetProbableHighestNonceIfNeededCalled func()
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState) error {
	return fdm.AddHeaderCalled(header, hash, state)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64, hash []byte) {
	fdm.RemoveHeadersCalled(nonce, hash)
}

func (fdm *ForkDetectorMock) CheckFork() (bool, uint64, []byte) {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return fdm.GetHighestFinalBlockNonceCalled()
}

func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetProbableHighestNonceIfNeeded() {
	fdm.ResetProbableHighestNonceIfNeededCalled()
}
