package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled      func(hash []byte) error
	RequestDataFromHashArrayCalled func(hashes [][]byte) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) ([]byte, error)
	GetMiniBlocksCalled            func(hashes [][]byte) block.MiniBlockSlice
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte) error {
	return hrm.RequestDataFromHashCalled(hash)
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHashArray(hashes [][]byte) error {
	return hrm.RequestDataFromHashArrayCalled(hashes)
}

func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P) ([]byte, error) {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *MiniBlocksResolverMock) GetMiniBlocks(hashes [][]byte) block.MiniBlockSlice {
	return hrm.GetMiniBlocksCalled(hashes)
}
