package disabled

// BlockTracker -
type BlockTracker struct {
}

// IsShardStuck -
func (b *BlockTracker) IsShardStuck(_ uint32) bool {
	return false
}

// IsInterfaceNil -
func (b *BlockTracker) IsInterfaceNil() bool {
	return b == nil
}
