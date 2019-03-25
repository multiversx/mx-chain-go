package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	RemoveHeadersCalled         func(nonce uint64)
	AddHeaderCalled             func(header *block.Header, hash []byte, isProcessed bool) error
	RemoveProcessedHeaderCalled func(nonce uint64) error
	CheckForkCalled             func() bool
}

func (f *ForkDetectorMock) RemoveHeaders(nonce uint64) {
	f.RemoveHeadersCalled(nonce)
}

// AddHeader is a mock implementation for AddHeader
func (f *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isProcessed bool) error {
	return f.AddHeaderCalled(header, hash, isProcessed)
}

// RemoveHeader is a mock implementation for RemoveHeader
func (f *ForkDetectorMock) RemoveProcessedHeader(nonce uint64) error {
	return f.RemoveProcessedHeaderCalled(nonce)
}

// CheckFork is a mock implementation for CheckFork
func (f *ForkDetectorMock) CheckFork() bool {
	return f.CheckForkCalled()
}
