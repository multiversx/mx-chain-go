package types

import "time"

// Block is a structure containing all the fields that need
//  to be saved for a block. It has all the default fields
//  plus some extra information for ease of search and filter
type Block struct {
	Nonce                 uint64        `json:"nonce"`
	Round                 uint64        `json:"round"`
	Epoch                 uint32        `json:"epoch"`
	Hash                  string        `json:"-"`
	MiniBlocksHashes      []string      `json:"miniBlocksHashes"`
	NotarizedBlocksHashes []string      `json:"notarizedBlocksHashes"`
	Proposer              uint64        `json:"proposer"`
	Validators            []uint64      `json:"validators"`
	PubKeyBitmap          string        `json:"pubKeyBitmap"`
	Size                  int64         `json:"size"`
	SizeTxs               int64         `json:"sizeTxs"`
	Timestamp             time.Duration `json:"timestamp"`
	StateRootHash         string        `json:"stateRootHash"`
	PrevHash              string        `json:"prevHash"`
	ShardID               uint32        `json:"shardId"`
	TxCount               uint32        `json:"txCount"`
	AccumulatedFees       string        `json:"accumulatedFees"`
	DeveloperFees         string        `json:"developerFees"`
	EpochStartBlock       bool          `json:"epochStartBlock"`
	SearchOrder           uint64        `json:"searchOrder"`
}

// Miniblock is a structure containing miniblock information
type Miniblock struct {
	Hash              string        `json:"-"`
	SenderShardID     uint32        `json:"senderShard"`
	ReceiverShardID   uint32        `json:"receiverShard"`
	SenderBlockHash   string        `json:"senderBlockHash"`
	ReceiverBlockHash string        `json:"receiverBlockHash"`
	Type              string        `json:"type"`
	Timestamp         time.Duration `json:"timestamp"`
}
