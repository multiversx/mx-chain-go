package external

// RecentBlock is the entity used to hold relevant info about a recent block (notarized)
type RecentBlock struct {
	ShardID        uint32
	Nonce          uint64
	Hash           []byte
	PrevHash       []byte
	StateRootHash  []byte
	ProposerPubKey []byte
	BlockSize      int64
	Timestamp      uint64
	TxCount        uint32
}
