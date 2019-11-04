package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type ForkDetectorMock struct {
	AddHeaderCalled                         func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error
	RemoveHeadersCalled                     func(nonce uint64, hash []byte)
	CheckForkCalled                         func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled         func() uint64
	ProbableHighestNonceCalled              func() uint64
	ResetProbableHighestNonceIfNeededCalled func()
	ResetProbableHighestNonceCalled         func()
	ResetForkCalled                         func()
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
	return fdm.AddHeaderCalled(header, hash, state, finalHeaders, finalHeadersHashes)
}

func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64, hash []byte) {
	fdm.RemoveHeadersCalled(nonce, hash)
}

func (fdm *ForkDetectorMock) CheckFork() *process.ForkInfo {
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

func (fdm *ForkDetectorMock) ResetProbableHighestNonce() {
	fdm.ResetProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetFork() {
	fdm.ResetForkCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorMock) IsInterfaceNil() bool {
	if fdm == nil {
		return true
	}
	return false
}
