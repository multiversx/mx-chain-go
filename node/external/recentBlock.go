package external

// RecentBlock is the entity used to hold relevant info about a recent block (notarized)
type RecentBlock struct {
	ShardId        uint32
	Nonce          uint64
	Hash           []byte
	PrevHash       []byte
	StateRootHash  []byte
	ProposerPubKey []byte
	PubKeysBitmap  []byte
	BlockSize      int64
	TimeStamp      uint64
	TxCount        uint32
}
