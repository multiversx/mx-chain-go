package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

// MiniBlocksResolverStub -
type MiniBlocksResolverStub struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	GetMiniBlocksCalled            func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPoolCalled    func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
}

// RequestDataFromHash -
func (hrm *MiniBlocksResolverStub) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

// RequestDataFromHashArray -
func (hrm *MiniBlocksResolverStub) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return hrm.RequestDataFromHashArrayCalled(hashes, epoch)
}

// ProcessReceivedMessage -
func (hrm *MiniBlocksResolverStub) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

// GetMiniBlocks -
func (hrm *MiniBlocksResolverStub) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return hrm.GetMiniBlocksCalled(hashes)
}

// GetMiniBlocksFromPool -
func (hrm *MiniBlocksResolverStub) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return hrm.GetMiniBlocksFromPoolCalled(hashes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *MiniBlocksResolverStub) IsInterfaceNil() bool {
	return hrm == nil
}
