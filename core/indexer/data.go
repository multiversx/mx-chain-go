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
	Value         *big.Int      `json:"value"`
	Receiver      string        `json:"receiver"`
	Sender        string        `json:"sender"`
	ReceiverShard uint32        `json:"receiverShard"`
	SenderShard   uint32        `json:"senderShard"`
	GasPrice      uint64        `json:"gasPrice"`
	GasLimit      uint64        `json:"gasLimit"`
	Data          string        `json:"data"`
	Signature     string        `json:"signature"`
	Timestamp     time.Duration `json:"timestamp"`
	Status        string        `json:"status"`
}

// Block is a structure containing all the fields that need
//  to be saved for a block. It has all the default fields
//  plus some extra information for ease of search and filter
type Block struct {
	Nonce         uint64        `json:"nonce"`
	ShardID       uint32        `json:"shardId"`
	Hash          string        `json:"hash"`
	Proposer      string        `json:"proposer"`
	Validators    []string      `json:"validators"`
	PubKeyBitmap  string        `json:"pubKeyBitmap"`
	Size          int64         `json:"size"`
	Timestamp     time.Duration `json:"timestamp"`
	TxCount       uint32        `json:"txCount"`
	StateRootHash string        `json:"stateRootHash"`
	PrevHash      string        `json:"prevHash"`
}

// TPS is a structure containing all the fields that need to
//  be saved for a shard statistic in the database
type TPS struct {
	LiveTPS               float64  `json:"liveTPS"`
	PeakTPS               float64  `json:"peakTPS"`
	NrOfShards            uint32   `json:"nrOfShards"`
	NrOfNodes             uint32   `json:"nrOfNodes"`
	BlockNumber           uint64   `json:"blockNumber"`
	RoundNumber           uint64   `json:"roundNumber"`
	RoundTime             uint64   `json:"roundTime"`
	AverageBlockTxCount   *big.Int `json:"averageBlockTxCount"`
	LastBlockTxCount      uint32   `json:"lastBlockTxCount"`
	TotalProcessedTxCount *big.Int `json:"totalProcessedTxCount"`
	ShardID               uint32   `json:"shardID"`
	AverageTPS            *big.Int `json:"averageTPS"`
	CurrentBlockNonce     uint64   `json:"currentBlockNonce"`
}
