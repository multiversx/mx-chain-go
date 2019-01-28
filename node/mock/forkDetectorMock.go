package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
)

// ForkDetectorMock is a mock implementation for the ForkDetector interface
type ForkDetectorMock struct {
	AddHeaderCalled func(header *block.Header, hash []byte, isReceived bool) error
	RemoveHeaderCalled func(nonce uint64)
	CheckForkCalled func() bool
}

// AddHeader is a mock implementation for AddHeader
func (f *ForkDetectorMock) AddHeader(header *block.Header, hash []byte, isReceived bool) error {
	return f.AddHeaderCalled(header, hash, isReceived)
}

// RemoveHeader is a mock implementation for RemoveHeader
func (f *ForkDetectorMock) RemoveHeader(nonce uint64) {
	f.RemoveHeaderCalled(nonce)
}

// CheckFork is a mock implementation for CheckFork
func (f *ForkDetectorMock) CheckFork() bool {
	return f.CheckForkCalled()
}
