package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	AddHeaderCalled                  func(header *block.Header, hash []byte, isProcessed bool) error
	RemoveHeadersCalled              func(nonce uint64)
	CheckForkCalled                  func() (bool, uint64)
	GetHighestSignedBlockNonceCalled func() uint64
	GetHighestFinalBlockNonceCalled  func() uint64
}

// AddHeader is a mock implementation for AddHeader
func (f *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return f.AddHeaderCalled(header, hash, isProcessed)
}

// RemoveHeaders is a mock implementation for RemoveHeaders
func (f *ForkDetectorMock) RemoveHeaders(nonce uint64) {
	f.RemoveHeadersCalled(nonce)
}

// CheckFork is a mock implementation for CheckFork
func (f *ForkDetectorMock) CheckFork() (bool, uint64) {
	return f.CheckForkCalled()
}

// GetHighestSignedBlockNonce is a mock implementation for GetHighestSignedBlockNonce
func (f *ForkDetectorMock) GetHighestSignedBlockNonce() uint64 {
	return f.GetHighestSignedBlockNonceCalled()
}

// GetHighestFinalBlockNonce is a mock implementation for GetHighestFinalBlockNonce
func (f *ForkDetectorMock) GetHighestFinalBlockNonce() uint64 {
	return f.GetHighestFinalBlockNonceCalled()
}
