package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ForkDetectorMock -
type ForkDetectorMock struct {
	AddHeaderCalled                 func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error
	RemoveHeaderCalled              func(nonce uint64, hash []byte)
	CheckForkCalled                 func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled func() uint64
	GetHighestFinalBlockHashCalled  func() []byte
	ProbableHighestNonceCalled      func() uint64
	ResetForkCalled                 func()
	GetNotarizedHeaderHashCalled    func(nonce uint64) []byte
	SetRollBackNonceCalled          func(nonce uint64)
	RestoreToGenesisCalled          func()
	ResetProbableHighestNonceCalled func()
	SetFinalToLastCheckpointCalled  func()
}

// RestoreToGenesis -
func (fdm *ForkDetectorMock) RestoreToGenesis() {
	fdm.RestoreToGenesisCalled()
}

// AddHeader -
func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
	return fdm.AddHeaderCalled(header, hash, state, selfNotarizedHeaders, selfNotarizedHeadersHashes)
}

// RemoveHeader -
func (fdm *ForkDetectorMock) RemoveHeader(nonce uint64, hash []byte) {
	fdm.RemoveHeaderCalled(nonce, hash)
}

// CheckFork -
func (fdm *ForkDetectorMock) CheckFork() *process.ForkInfo {
	return fdm.CheckForkCalled()
}

// GetHighestFinalBlockNonce -
func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	if fdm.GetHighestFinalBlockNonceCalled != nil {
		return fdm.GetHighestFinalBlockNonceCalled()
	}
	return 0
}

// GetHighestFinalBlockHash -
func (fdm *ForkDetectorMock) GetHighestFinalBlockHash() []byte {
	return fdm.GetHighestFinalBlockHashCalled()
}

// ProbableHighestNonce -
func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}

// SetRollBackNonce -
func (fdm *ForkDetectorMock) SetRollBackNonce(nonce uint64) {
	if fdm.SetRollBackNonceCalled != nil {
		fdm.SetRollBackNonceCalled(nonce)
	}
}

// ResetFork -
func (fdm *ForkDetectorMock) ResetFork() {
	fdm.ResetForkCalled()
}

// GetNotarizedHeaderHash -
func (fdm *ForkDetectorMock) GetNotarizedHeaderHash(nonce uint64) []byte {
	return fdm.GetNotarizedHeaderHashCalled(nonce)
}

// ResetProbableHighestNonce -
func (fdm *ForkDetectorMock) ResetProbableHighestNonce() {
	if fdm.ResetProbableHighestNonceCalled != nil {
		fdm.ResetProbableHighestNonceCalled()
	}
}

// SetFinalToLastCheckpoint -
func (fdm *ForkDetectorMock) SetFinalToLastCheckpoint() {
	if fdm.SetFinalToLastCheckpointCalled != nil {
		fdm.SetFinalToLastCheckpointCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorMock) IsInterfaceNil() bool {
	return fdm == nil
}
