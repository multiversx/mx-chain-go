package update

// GetLastPostMbs -
func GetLastPostMbs(shardIDs []uint32, mapPostMbs map[uint32][]*MbInfo) []*MbInfo {
	return getLastPostMbs(shardIDs, mapPostMbs)
}
