package disabled

// BlockTracker implements the BlockTracker interface but does nothing as it is disabled
type BlockTracker struct {
}

// IsShardStuck returns false as this is a disabled implementation
func (b *BlockTracker) IsShardStuck(_ uint32) bool {
	return false
}

// IsInterfaceNil returns true if underlying object is nil
func (b *BlockTracker) IsInterfaceNil() bool {
	return b == nil
}
