package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

type ForkDetectorMock struct {
	AddHeaderCalled                  func(header *block.Header, hash []byte, isReceived bool) error
	ResetProcessedHeaderCalled       func(nonce uint64) error
	CheckForkCalled                  func() bool
	GetHighestSignedBlockNonceCalled func() uint64
}

func (fdm *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isReceived bool) error {
	return fdm.AddHeaderCalled(header, hash, isReceived)
}

func (fdm *ForkDetectorMock) ResetProcessedHeader(nonce uint64) error {
	return fdm.ResetProcessedHeaderCalled(nonce)
}

func (fdm *ForkDetectorMock) CheckFork() bool {
	return fdm.CheckForkCalled()
}

func (fdm *ForkDetectorMock) GetHighestSignedBlockNonce() uint64 {
	return fdm.GetHighestSignedBlockNonceCalled()
}
