package types

import (
	"math/big"
	"time"
)

//ValidatorsPublicKeys is a structure containing fields for validators public keys
type ValidatorsPublicKeys struct {
	PublicKeys []string `json:"publicKeys"`
}

// TPS is a structure containing all the fields that need to
//  be saved for a shard statistic in the database
type TPS struct {
	LiveTPS               float64  `json:"liveTPS"`
	PeakTPS               float64  `json:"peakTPS"`
	BlockNumber           uint64   `json:"blockNumber"`
	RoundNumber           uint64   `json:"roundNumber"`
	RoundTime             uint64   `json:"roundTime"`
	AverageBlockTxCount   *big.Int `json:"averageBlockTxCount"`
	TotalProcessedTxCount *big.Int `json:"totalProcessedTxCount"`
	AverageTPS            *big.Int `json:"averageTPS"`
	CurrentBlockNonce     uint64   `json:"currentBlockNonce"`
	NrOfShards            uint32   `json:"nrOfShards"`
	NrOfNodes             uint32   `json:"nrOfNodes"`
	LastBlockTxCount      uint32   `json:"lastBlockTxCount"`
	ShardID               uint32   `json:"shardID"`
}

// KibanaResponse -
type KibanaResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"-"`
	Rating    float32 `json:"rating"`
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
}

// EpochInfo holds the information about epoch
type EpochInfo struct {
	AccumulatedFees string `json:"accumulatedFees"`
	DeveloperFees   string `json:"developerFees"`
}
