package state

// peerAccountsDB will save and synchronize data from peer processor, plus will synchronize with nodesCoordinator
type peerAccountsDB struct {
	*AccountsDB
}
