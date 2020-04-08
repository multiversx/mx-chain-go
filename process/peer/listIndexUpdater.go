package peer

// ListIndexUpdater will handle the updating of list type and the index for a peer
type ListIndexUpdater struct {
	updateListAndIndex func(pubKey string, shardID uint32, list string, index uint32) error
}

// UpdateListAndIndex will update the list and the index for a given peer
func (liu *ListIndexUpdater) UpdateListAndIndex(pubKey string, shardID uint32, list string, index uint32) error {
	return liu.updateListAndIndex(pubKey, shardID, list, index)
}

// IsInterfaceNil checks if the underlying object is nil
func (liu *ListIndexUpdater) IsInterfaceNil() bool {
	return liu == nil
}
