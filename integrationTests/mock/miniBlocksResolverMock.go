package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled      func(hash []byte, epoch uint32) error
	RequestDataFromHashArrayCalled func(hashes [][]byte, epoch uint32) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	GetMiniBlocksCalled            func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
	GetMiniBlocksFromPoolCalled    func(hashes [][]byte) (block.MiniBlockSlice, [][]byte)
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte, epoch uint32) error {
	return hrm.RequestDataFromHashCalled(hash, epoch)
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHashArray(hashes [][]byte, epoch uint32) error {
	return hrm.RequestDataFromHashArrayCalled(hashes, epoch)
}

func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ p2p.PeerID) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *MiniBlocksResolverMock) GetMiniBlocks(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return hrm.GetMiniBlocksCalled(hashes)
}

func (hrm *MiniBlocksResolverMock) GetMiniBlocksFromPool(hashes [][]byte) (block.MiniBlockSlice, [][]byte) {
	return hrm.GetMiniBlocksFromPoolCalled(hashes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *MiniBlocksResolverMock) IsInterfaceNil() bool {
	return hrm == nil
}
