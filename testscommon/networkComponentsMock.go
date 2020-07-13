package testscommon

import (
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

type NetworkComponentsMock struct {
	Messenger       p2p.Messenger
	InputAntiFlood  process.P2PAntifloodHandler
	OutputAntiFlood process.P2PAntifloodHandler
	PeerBlackList   process.PeerBlackListCacher
}

// NetworkMessenger -
func (ncm *NetworkComponentsMock) NetworkMessenger() p2p.Messenger {
	return ncm.Messenger
}

// InputAntiFloodHandler -
func (ncm *NetworkComponentsMock) InputAntiFloodHandler() process.P2PAntifloodHandler {
	return ncm.InputAntiFlood
}

// OutputAntiFloodHandler -
func (ncm *NetworkComponentsMock) OutputAntiFloodHandler() process.P2PAntifloodHandler {
	return ncm.OutputAntiFlood
}

// PeerBlackListHandler -
func (ncm *NetworkComponentsMock) PeerBlackListHandler() process.PeerBlackListCacher {
	return ncm.PeerBlackList
}

// IsInterfaceNil -
func (ncm *NetworkComponentsMock) IsInterfaceNil() bool {
	return ncm == nil
}
