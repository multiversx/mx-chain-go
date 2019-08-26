package mock

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/p2p"
)

type MiniBlocksResolverMock struct {
	RequestDataFromHashCalled      func(hash []byte) error
	RequestDataFromHashArrayCalled func(hashes [][]byte) error
	ProcessReceivedMessageCalled   func(message p2p.MessageP2P) error
	GetMiniBlocksCalled            func(hashes [][]byte) block.MiniBlockSlice
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHash(hash []byte) error {
	return hrm.RequestDataFromHashCalled(hash)
}

func (hrm *MiniBlocksResolverMock) RequestDataFromHashArray(hashes [][]byte) error {
	return hrm.RequestDataFromHashArrayCalled(hashes)
}

func (hrm *MiniBlocksResolverMock) ProcessReceivedMessage(message p2p.MessageP2P, _ func(buffToSend []byte)) error {
	return hrm.ProcessReceivedMessageCalled(message)
}

func (hrm *MiniBlocksResolverMock) GetMiniBlocks(hashes [][]byte) block.MiniBlockSlice {
	return hrm.GetMiniBlocksCalled(hashes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (hrm *MiniBlocksResolverMock) IsInterfaceNil() bool {
	if hrm == nil {
		return true
	}
	return false
}
