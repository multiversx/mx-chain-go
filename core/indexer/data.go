package indexer

import (
	"math/big"
	"time"
)

// Transaction is a structure containing all the fields that need
//  to be saved for a transaction. It has all the default fields
//  plus some extra information for ease of search and filter
type Transaction struct {
	Hash          string        `json:"hash"`
	MBHash        string        `json:"miniBlockHash"`
	BlockHash     string        `json:"blockHash"`
	Nonce         uint64        `json:"nonce"`
	Round         uint64        `json:"round"`
	Value         string        `json:"value"`
	Receiver      string        `json:"receiver"`
	Sender        string        `json:"sender"`
	ReceiverShard uint32        `json:"receiverShard"`
	SenderShard   uint32        `json:"senderShard"`
	GasPrice      uint64        `json:"gasPrice"`
	GasLimit      uint64        `json:"gasLimit"`
	Data          []byte        `json:"data"`
	Signature     string        `json:"signature"`
	Timestamp     time.Duration `json:"timestamp"`
	Status        string        `json:"status"`
}

// Block is a structure containing all the fields that need
//  to be saved for a block. It has all the default fields
//  plus some extra information for ease of search and filter
type Block struct {
	Nonce         uint64        `json:"nonce"`
	Round         uint64        `json:"round"`
	Hash          string        `json:"hash"`
	Proposer      uint64        `json:"proposer"`
	Validators    []uint64      `json:"validators"`
	PubKeyBitmap  string        `json:"pubKeyBitmap"`
	Size          int64         `json:"size"`
	Timestamp     time.Duration `json:"timestamp"`
	StateRootHash string        `json:"stateRootHash"`
	PrevHash      string        `json:"prevHash"`
	ShardID       uint32        `json:"shardId"`
	TxCount       uint32        `json:"txCount"`
}

//ValidatorsPublicKeys is a structure containing fields for validators public keys
type ValidatorsPublicKeys struct {
	PublicKeys []string `json:"publicKeys"`
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"-"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
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
