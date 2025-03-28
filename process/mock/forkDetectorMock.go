package mock

import (
	"github.com/multiversx/mx-chain-core-go/data"

	"github.com/multiversx/mx-chain-go/process"
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
	ReceivedProofCalled             func(proof data.HeaderProofHandler)
	AddCheckpointCalled             func(nonce uint64, round uint64, hash []byte)
}

// RestoreToGenesis -
func (fdm *ForkDetectorMock) RestoreToGenesis() {
	fdm.RestoreToGenesisCalled()
}

// AddHeader -
func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
	if fdm.AddHeaderCalled != nil {
		return fdm.AddHeaderCalled(header, hash, state, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}

	return nil
}

// RemoveHeader -
func (fdm *ForkDetectorMock) RemoveHeader(nonce uint64, hash []byte) {
	if fdm.RemoveHeaderCalled != nil {
		fdm.RemoveHeaderCalled(nonce, hash)
	}
}

// CheckFork -
func (fdm *ForkDetectorMock) CheckFork() *process.ForkInfo {
	if fdm.CheckForkCalled != nil {
		return fdm.CheckForkCalled()
	}

	return nil
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
	if fdm.GetHighestFinalBlockHashCalled != nil {
		return fdm.GetHighestFinalBlockHashCalled()
	}

	return nil
}

// ProbableHighestNonce -
func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	if fdm.ProbableHighestNonceCalled != nil {
		return fdm.ProbableHighestNonceCalled()
	}

	return 0
}

// SetRollBackNonce -
func (fdm *ForkDetectorMock) SetRollBackNonce(nonce uint64) {
	if fdm.SetRollBackNonceCalled != nil {
		fdm.SetRollBackNonceCalled(nonce)
	}
}

// ResetFork -
func (fdm *ForkDetectorMock) ResetFork() {
	if fdm.ResetForkCalled != nil {
		fdm.ResetForkCalled()
	}
}

// GetNotarizedHeaderHash -
func (fdm *ForkDetectorMock) GetNotarizedHeaderHash(nonce uint64) []byte {
	if fdm.GetNotarizedHeaderHashCalled != nil {
		return fdm.GetNotarizedHeaderHashCalled(nonce)
	}

	return nil
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

// ReceivedProof -
func (fdm *ForkDetectorMock) ReceivedProof(proof data.HeaderProofHandler) {
	if fdm.ReceivedProofCalled != nil {
		fdm.ReceivedProofCalled(proof)
	}
}

// AddCheckpoint -
func (fdm *ForkDetectorMock) AddCheckpoint(nonce uint64, round uint64, hash []byte) {
	if fdm.AddCheckpointCalled != nil {
		fdm.AddCheckpointCalled(nonce, round, hash)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorMock) IsInterfaceNil() bool {
	return fdm == nil
}
