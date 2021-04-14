package api

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/data/transaction"
)

// Block represents the structure for block that is returned by api routes
type Block struct {
	Nonce                  uint64            `json:"nonce"`
	Round                  uint64            `json:"round"`
	Hash                   string            `json:"hash"`
	PrevBlockHash          string            `json:"prevBlockHash"`
	Epoch                  uint32            `json:"epoch"`
	Shard                  uint32            `json:"shard"`
	NumTxs                 uint32            `json:"numTxs"`
	NotarizedBlocks        []*NotarizedBlock `json:"notarizedBlocks,omitempty"`
	MiniBlocks             []*MiniBlock      `json:"miniBlocks,omitempty"`
	Timestamp              time.Duration     `json:"timestamp,omitempty"`
	AccumulatedFees        string            `json:"accumulatedFees,omitempty"`
	DeveloperFees          string            `json:"developerFees,omitempty"`
	AccumulatedFeesInEpoch string            `json:"accumulatedFeesInEpoch,omitempty"`
	DeveloperFeesInEpoch   string            `json:"developerFeesInEpoch,omitempty"`
	Status                 string            `json:"status,omitempty"`
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

// StakeValues is the structure that contains the total staked value and the total top up value
type StakeValues struct {
	TotalStaked *big.Int
	TopUp       *big.Int
}

// DirectStakedValue holds the total staked value for an address
type DirectStakedValue struct {
	Address string `json:"address"`
	Staked  string `json:"staked"`
	TopUp   string `json:"topUp"`
	Total   string `json:"total"`
}

// DelegatedValue holds the value and the delegation system SC address
type DelegatedValue struct {
	DelegationScAddress string `json:"delegationScAddress"`
	Value               string `json:"value"`
}

// Delegator holds the delegator address and the slice of delegated values
type Delegator struct {
	DelegatorAddress string            `json:"delegatorAddress"`
	DelegatedTo      []*DelegatedValue `json:"delegatedTo"`
	Total            string            `json:"total"`
	TotalAsBigInt    *big.Int          `json:"-"`
}
