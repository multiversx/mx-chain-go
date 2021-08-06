package syncer

// SyncStatisticsHandler defines the methods for a component able to store the sync statistics for a trie
type SyncStatisticsHandler interface {
	Reset()
	AddNumReceived(value int)
	AddNumLarge(value int)
	SetNumMissing(rootHash []byte, value int)
	NumReceived() int
	NumLarge() int
	NumMissing() int
	IsInterfaceNil() bool
}
