package state

import "github.com/ElrondNetwork/elrond-go/sharding"

// peerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type peerAccountsDB struct {
	*AccountsDB

	shardCoordinator sharding.Coordinator
}
