package libp2p

// nilBlacklistHandler is a mock implementation of BlacklistHandler that does not manage black listed keys
// (all keys [peers] are whitelisted)
type nilBlacklistHandler struct {
}

// Has outputs false (all peers are white listed)
func (nbh *nilBlacklistHandler) Has(_ string) bool {
	return false
}

// IsInterfaceNil returns true if there is no value under the interface
func (nbh *nilBlacklistHandler) IsInterfaceNil() bool {
	return nbh == nil
}
