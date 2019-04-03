package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type ForkDetectorMock struct {
	AddHeaderCalled                    func(header *block.Header, hash []byte, isProcessed bool) error
	RemoveProcessedHeaderCalled        func(nonce uint64) error
	CheckForkCalled                    func() bool
	GetHighestSignedBlockNonceCalled   func() uint64
	GetHighestFinalityBlockNonceCalled func() uint64
}

func (fdm *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return fdm.AddHeaderCalled(header, hash, isProcessed)
}

func (fdm *ForkDetectorMock) RemoveProcessedHeader(nonce uint64) error {
	return fdm.RemoveProcessedHeaderCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() bool {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestSignedBlockNonce() uint64 {
	return fdm.GetHighestSignedBlockNonceCalled()
}

func (fdm *ForkDetectorMock) GetHighestFinalityBlockNonce() uint64 {
	return fdm.GetHighestFinalityBlockNonceCalled()
}
