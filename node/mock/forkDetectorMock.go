package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	AddHeaderCalled                         func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeader data.HeaderHandler, finalHeaderHash []byte) error
	RemoveHeadersCalled                     func(nonce uint64, hash []byte)
	CheckForkCalled                         func() (bool, uint64, []byte)
	GetHighestFinalBlockNonceCalled         func() uint64
	ProbableHighestNonceCalled              func() uint64
	ResetProbableHighestNonceIfNeededCalled func()
}

// AddHeader is a mock implementation for AddHeader
func (f *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeader data.HeaderHandler, finalHeaderHash []byte) error {
	return f.AddHeaderCalled(header, hash, state, finalHeader, finalHeaderHash)
}

// RemoveHeaders is a mock implementation for RemoveHeaders
func (f *ForkDetectorMock) RemoveHeaders(nonce uint64, hash []byte) {
	f.RemoveHeadersCalled(nonce, hash)
}

// CheckFork is a mock implementation for CheckFork
func (f *ForkDetectorMock) CheckFork() (bool, uint64, []byte) {
	return f.CheckForkCalled()
}

// GetHighestFinalBlockNonce is a mock implementation for GetHighestFinalBlockNonce
func (f *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return f.GetHighestFinalBlockNonceCalled()
}

// ProbableHighestNonce is a mock implementation for GetProbableHighestNonce
func (f *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return f.ProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetProbableHighestNonceIfNeeded() {
	fdm.ResetProbableHighestNonceIfNeededCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorMock) IsInterfaceNil() bool {
	if fdm == nil {
		return true
	}
	return false
}
