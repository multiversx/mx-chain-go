package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type ForkDetectorMock struct {
	AddHeaderCalled                       func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error
	RemoveHeaderCalled                    func(nonce uint64, hash []byte)
	CheckForkCalled                       func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled       func() uint64
	GetHighestFinalBlockHashCalled        func() []byte
	ProbableHighestNonceCalled            func() uint64
	ResetForkCalled                       func()
	GetNotarizedHeaderHashCalled          func(nonce uint64) []byte
	SetRollBackNonceCalled                func(nonce uint64)
	RestoreFinalCheckPointToGenesisCalled func()
}

func (fdm *ForkDetectorMock) RestoreFinalCheckPointToGenesis() {
	fdm.RestoreFinalCheckPointToGenesisCalled()
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, selfNotarizedHeaders []data.HeaderHandler, selfNotarizedHeadersHashes [][]byte) error {
	return fdm.AddHeaderCalled(header, hash, state, selfNotarizedHeaders, selfNotarizedHeadersHashes)
}

func (fdm *ForkDetectorMock) RemoveHeader(nonce uint64, hash []byte) {
	fdm.RemoveHeaderCalled(nonce, hash)
}

func (fdm *ForkDetectorMock) CheckFork() *process.ForkInfo {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return fdm.GetHighestFinalBlockNonceCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalBlockHash() []byte {
	return fdm.GetHighestFinalBlockHashCalled()
}

func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) SetRollBackNonce(nonce uint64) {
	if fdm.SetRollBackNonceCalled != nil {
		fdm.SetRollBackNonceCalled(nonce)
	}
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
