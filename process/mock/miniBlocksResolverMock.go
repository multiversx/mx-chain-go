package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
)

type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled    func(hash []byte) error
	ProcessReceivedMessageCalled func(message p2p.MessageP2P) error
	GetMiniBlocksCalled func(hashes [][]byte) []*block.MiniBlock
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte) error {
	return hrm.RequestDataFromHashCalled(hash)
}

func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *MiniBlocksResolverMock) GetMiniBlocks(hashes [][]byte) []*block.MiniBlock {
	return hrm.GetMiniBlocksCalled(hashes)
}
