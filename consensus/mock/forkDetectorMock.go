package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/process"
)

type ForkDetectorMock struct {
	AddHeaderCalled                       func(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, notarizedHeaders []data.HeaderHandler, notarizedHeadersHashes [][]byte) error
	RemoveHeadersCalled                   func(nonce uint64, hash []byte)
	CheckForkCalled                       func() *process.ForkInfo
	GetHighestFinalBlockNonceCalled       func() uint64
	GetHighestFinalBlockHashCalled        func() []byte
	ProbableHighestNonceCalled            func() uint64
	ResetProbableHighestNonceCalled       func()
	ResetForkCalled                       func()
	GetNotarizedHeaderHashCalled          func(nonce uint64) []byte
	SetForkNonceCalled                    func(nonce uint64)
	RestoreFinalCheckPointToGenesisCalled func()
}

func (fdm *ForkDetectorMock) RestoreFinalCheckPointToGenesis() {
	fdm.RestoreFinalCheckPointToGenesisCalled()
}

func (fdm *ForkDetectorMock) AddHeader(header data.HeaderHandler, hash []byte, state process.BlockHeaderState, notarizedHeaders []data.HeaderHandler, notarizedHeadersHashes [][]byte) error {
	return fdm.AddHeaderCalled(header, hash, state, notarizedHeaders, notarizedHeadersHashes)
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

func (fdm *ForkDetectorMock) GetHighestFinalBlockHash() []byte {
	return fdm.GetHighestFinalBlockHashCalled()
}

func (fdm *ForkDetectorMock) ProbableHighestNonce() uint64 {
	return fdm.ProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) ResetProbableHighestNonce() {
	fdm.ResetProbableHighestNonceCalled()
}

func (fdm *ForkDetectorMock) SetForkNonce(nonce uint64) {
	if fdm.SetForkNonceCalled != nil {
		fdm.SetForkNonceCalled(nonce)
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
