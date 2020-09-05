package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

// ForkDetectorStub is a mock implementation for the ForkDetector interface
type ForkDetectorStub struct {
	AddHeaderCalled                 func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error
	RemoveHeaderCalled              func(nonce uint64, hash []byte)
	CheckForkCalled                 func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled func() uint64
	GetHighestFinalBlockHashCalled  func() []byte
	ProbableHighestNonceCalled      func() uint64
	ResetForkCalled                 func()
	GetNotarizedHeaderHashCalled    func(nonce uint64) []byte
	RestoreToGenesisCalled          func()
	SetRollBackNonceCalled          func(nonce uint64)
	ResetProbableHighestNonceCalled func()
	SetFinalToLastCheckpointCalled  func()
}

// RestoreToGenesis -
func (fdm *ForkDetectorStub) RestoreToGenesis() {
	if fdm.RestoreToGenesisCalled != nil {
		fdm.RestoreToGenesisCalled()
	}
}

// AddHeader is a mock implementation for AddHeader
func (fdm *ForkDetectorStub) AddHeader(
	header data.HeaderHandler,
	hash []byte,
	state process.BlockHeaderState,
	selfNotarizedHeaders []data.HeaderHandler,
	selfNotarizedHeadersHashes [][]byte,
) error {
	if fdm.AddHeaderCalled != nil {
		return fdm.AddHeaderCalled(header, hash, state, selfNotarizedHeaders, selfNotarizedHeadersHashes)
	}
	return nil
}

// RemoveHeader is a mock implementation for RemoveHeader
func (fdm *ForkDetectorStub) RemoveHeader(nonce uint64, hash []byte) {
	if fdm.RemoveHeaderCalled != nil {
		fdm.RemoveHeaderCalled(nonce, hash)
	}
}

// CheckFork is a mock implementation for CheckFork
func (fdm *ForkDetectorStub) CheckFork() *process.ForkInfo {
	if fdm.CheckForkCalled != nil {
		return fdm.CheckForkCalled()
	}
	return &process.ForkInfo{}
}

// GetHighestFinalBlockNonce is a mock implementation for GetHighestFinalBlockNonce
func (fdm *ForkDetectorStub) GetHighestFinalBlockNonce() uint64 {
	if fdm.GetHighestFinalBlockNonceCalled != nil {
		return fdm.GetHighestFinalBlockNonceCalled()
	}
	return 0
}

// GetHighestFinalBlockHash -
func (fdm *ForkDetectorStub) GetHighestFinalBlockHash() []byte {
	return fdm.GetHighestFinalBlockHashCalled()
}

// ProbableHighestNonce is a mock implementation for ProbableHighestNonce
func (fdm *ForkDetectorStub) ProbableHighestNonce() uint64 {
	if fdm.ProbableHighestNonceCalled != nil {
		return fdm.ProbableHighestNonceCalled()
	}
	return 0
}

// SetRollBackNonce -
func (fdm *ForkDetectorStub) SetRollBackNonce(nonce uint64) {
	if fdm.SetRollBackNonceCalled != nil {
		fdm.SetRollBackNonceCalled(nonce)
	}
}

// ResetFork -
func (fdm *ForkDetectorStub) ResetFork() {
	if fdm.ResetForkCalled != nil {
		fdm.ResetForkCalled()
	}
}

// GetNotarizedHeaderHash -
func (fdm *ForkDetectorStub) GetNotarizedHeaderHash(nonce uint64) []byte {
	if fdm.GetNotarizedHeaderHashCalled != nil {
		return fdm.GetNotarizedHeaderHashCalled(nonce)
	}
	return nil
}

// ResetProbableHighestNonce -
func (fdm *ForkDetectorStub) ResetProbableHighestNonce() {
	if fdm.ResetProbableHighestNonceCalled != nil {
		fdm.ResetProbableHighestNonceCalled()
	}
}

// SetFinalToLastCheckpoint -
func (fdm *ForkDetectorStub) SetFinalToLastCheckpoint() {
	if fdm.SetFinalToLastCheckpointCalled != nil {
		fdm.SetFinalToLastCheckpointCalled()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (fdm *ForkDetectorStub) IsInterfaceNil() bool {
	return fdm == nil
}
