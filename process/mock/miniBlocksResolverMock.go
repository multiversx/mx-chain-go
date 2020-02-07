package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MiniBlocksResolverMock -
type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	GetMiniBlocksCalled            func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPoolCalled    func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
}

// RequestDataFromHash -
func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

// RequestDataFromHashArray -
func (hrm *MiniBlocksResolverMock) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return hrm.RequestDataFromHashArrayCalled(hashes, epoch)
}

// ProcessReceivedMessage -
func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

// GetMiniBlocks -
func (hrm *MiniBlocksResolverMock) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	if hrm.GetMiniBlocksCalled != nil {
		return hrm.GetMiniBlocksCalled(hashes)
	}
	return nil, nil
}

// GetMiniBlocksFromPool -
func (hrm *MiniBlocksResolverMock) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	if hrm.GetMiniBlocksFromPoolCalled != nil {
		return hrm.GetMiniBlocksFromPoolCalled(hashes)
	}
	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *MiniBlocksResolverMock) IsInterfaceNil() bool {
	if hrm == nil {
		return true
	}
	return false
}
