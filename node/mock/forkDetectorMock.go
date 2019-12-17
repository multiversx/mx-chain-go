package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	AddHeaderCalled                       func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error
	RemoveHeadersCalled                   func(nonce uint64, hash []byte)
	CheckForkCalled                       func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled       func() uint64
	ProbableHighestNonceCalled            func() uint64
	ResetProbableHighestNonceCalled       func()
	ResetForkCalled                       func()
	GetNotarizedHeaderHashCalled          func(nonce uint64) []byte
	RestoreFinalCheckPointToGenesisCalled func()
}

func (fdm *ForkDetectorMock) RestoreFinalCheckPointToGenesis() {
	fdm.RestoreFinalCheckPointToGenesisCalled()
}

// AddTrackedHeader is a mock implementation for AddTrackedHeader
func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, finalHeaders []data.HeaderHandler, finalHeadersHashes [][]byte) error {
	return fdm.AddHeaderCalled(header, hash, state, finalHeaders, finalHeadersHashes)
}

// RemoveHeaders is a mock implementation for RemoveHeaders
func (fdm *ForkDetectorMock) RemoveHeaders(nonce uint64, hash []byte) {
	fdm.RemoveHeadersCalled(nonce, hash)
}

// CheckFork is a mock implementation for CheckFork
func (fdm *ForkDetectorMock) CheckFork() *process.ForkInfo {
	return fdm.CheckForkCalled()
}

// GetHighestFinalBlockNonce is a mock implementation for GetHighestFinalBlockNonce
func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return fdm.GetHighestFinalBlockNonceCalled()
}

// ProbableHighestNonce is a mock implementation for GetProbableHighestNonce
func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetProbableHighestNonce() {
	fdm.ResetProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetFork() {
	fdm.ResetForkCalled()
}

func (fdm *ForkDetectorMock) GetNotarizedHeaderHash(nonce uint64) []byte {
	return fdm.GetNotarizedHeaderHashCalled(nonce)
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorMock) IsInterfaceNil() bool {
	return fdm == nil
}
