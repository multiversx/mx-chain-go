package external

import "math/big"

// TransactionStatus represents the status of a known transaction
type TransactionStatus byte

const (
	// TxUnknown represents an unknown transaction
	TxUnknown TransactionStatus = iota
	// TxPending represents a pending transaction (found only in datapools)
	TxPending
	// TxFailed represents a failed notarized transaction
	TxFailed
	// TxProcessing represents a transaction that is notarized only in sender shard
	TxProcessing
	// TxSucceed represents a succeeded transaction, fully notarized
	TxSucceed
)

// BlockHeader is the entity used to hold relevant info about a recent block (notarized)
type BlockHeader struct {
	ShardID        uint32
	Nonce          uint64
	Hash           []byte
	PrevHash       []byte
	StateRootHash  []byte
	ProposerPubKey []byte
	PubKeysBitmap  []byte
	BlockSize      int64
	Timestamp      uint64
	TxCount        uint32
}

// TransactionInfo is the entity used to hold relevant info about a transaction
type TransactionInfo struct {
	Nonce             uint64
	SenderAddress     []byte
	SenderShard       uint32
	ReceiverAddress   []byte
	ReceiverShard     uint32
	MiniblockHash     []byte
	SenderBlockHash   []byte
	ReceiverBlockHash []byte
	Balance           *big.Int
	Data              []byte
	Status            TransactionStatus
	Result            []byte
}

// ShardBlockInfo represents a notarized shard block with transactions
type ShardBlockInfo struct {
	BlockHeader
	Transactions []TransactionInfo
}
