package poolsCleaner

// BlockTracker defines the functionality for node to track the blocks which are received from network
type BlockTracker interface {
	GetNumPendingMiniBlocks(shardID uint32) uint32
	IsInterfaceNil() bool
}
