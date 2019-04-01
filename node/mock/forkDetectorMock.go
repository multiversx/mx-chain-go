package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	AddHeaderCalled                  func(header *block.Header, hash []byte, isProcessed bool) error
	ResetProcessedHeaderCalled       func(nonce uint64) error
	CheckForkCalled                  func() bool
	GetHighestSignedBlockNonceCalled func() uint64
}

// AddHeader is a mock implementation for AddHeader
func (f *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return f.AddHeaderCalled(header, hash, isProcessed)
}

// ResetProcessedHeader is a mock implementation for ResetProcessedHeader
func (f *ForkDetectorMock) ResetProcessedHeader(nonce uint64) error {
	return f.ResetProcessedHeaderCalled(nonce)
}

// CheckFork is a mock implementation for CheckFork
func (f *ForkDetectorMock) CheckFork() bool {
	return f.CheckForkCalled()
}

// GetHighestSignedBlockNonce is a mock implementation for GetHighestSignedBlockNonce
func (f *ForkDetectorMock) GetHighestSignedBlockNonce() uint64 {
	return f.GetHighestSignedBlockNonceCalled()
}
