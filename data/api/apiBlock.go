package api

import "github.com/ElrondNetwork/elrond-go/data/transaction"

// Block represents the structure for block that is returned by api routes
type Block struct {
	Nonce           uint64            `json:"nonce"`
	Round           uint64            `json:"round"`
	Hash            string            `json:"hash"`
	PrevBlockHash   string            `json:"prevBlockHash"`
	Epoch           uint32            `json:"epoch"`
	Shard           uint32            `json:"shard"`
	NumTxs          uint32            `json:"numTxs"`
	NotarizedBlocks []*NotarizedBlock `json:"notarizedBlocks,omitempty"`
	MiniBlocks      []*MiniBlock      `json:"miniBlocks,omitempty"`
}

// NotarizedBlock represents a notarized block
type NotarizedBlock struct {
	Hash  string `json:"hash"`
	Nonce uint64 `json:"nonce"`
	Shard uint32 `json:"shard"`
}

// MiniBlock represents the structure for a miniblock
type MiniBlock struct {
	Hash             string                              `json:"hash"`
	Type             string                              `json:"type"`
	SourceShard      uint32                              `json:"sourceShard"`
	DestinationShard uint32                              `json:"destinationShard"`
	Transactions     []*transaction.ApiTransactionResult `json:"transactions,omitempty"`
}
