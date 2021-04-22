package indexer

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// ArgsSaveBlockData will contains all information that are needed to save block data
type ArgsSaveBlockData struct {
	HeaderHash             []byte
	Body                   data.BodyHandler
	Header                 data.HeaderHandler
	SignersIndexes         []uint64
	NotarizedHeadersHashes []string
	TransactionsPool       *Pool
}

// Pool will holds all types of transaction
type Pool struct {
	Txs      map[string]data.TransactionHandler
	Scrs     map[string]data.TransactionHandler
	Rewards  map[string]data.TransactionHandler
	Invalid  map[string]data.TransactionHandler
	Receipts map[string]data.TransactionHandler
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string
	Rating    float32
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64
	SignersIndexes   []uint64
	BlockWasProposed bool
	ShardId          uint32
	Timestamp        time.Duration
}
