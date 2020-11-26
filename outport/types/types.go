package types

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/data"
)

// ArgsSaveBlocks will store arguments that are needed to call method SaveBlock from interface OutportHandler
type ArgsSaveBlocks struct {
	Body                   data.BodyHandler
	Header                 data.HeaderHandler
	TxsFromPool            map[string]data.TransactionHandler
	SignersIndexes         []uint64
	NotarizedHeadersHashes []string
	HeaderHash             []byte
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"publicKey"`
	Rating    float32 `json:"rating"`
}
