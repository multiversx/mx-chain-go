package indexer

import (
	"math/big"
)

// Transaction is a structure containing all the fields that need
//  to be saved for a transaction. It has all the default fields
//  plus some extra information for ease of search and filter
type Transaction struct {
	Hash          string   `json:"hash"`
	MBHash        string   `json:"miniBlockHash"`
	BlockHash     string   `json:"blockHash"`
	Nonce         uint64   `json:"nonce"`
	Value         *big.Int `json:"value"`
	Receiver      string   `json:"receiver"`
	Sender        string   `json:"sender"`
	ReceiverShard uint32   `json:"receiverShard"`
	SenderShard   uint32   `json:"senderShard"`
	GasPrice      uint64   `json:"gasPrice"`
	GasLimit      uint64   `json:"gasLimit"`
	Data          string   `json:"data"`
	Signature     string   `json:"signature"`
	Timestamp     uint64   `json:"timestamp"`
}

// Bock is a structure containing all the fields that need
//  to be saved for a block. It has all the default fields
//  plus some extra information for ease of search and filter
type Block struct {
	Nonce         uint64   `json:"nonce"`
	ShardID       uint32   `json:"shardId"`
	Hash          string   `json:"hash"`
	Proposer      string   `json:"proposer"`
	Validators    []string `json:"validators"`
	PubKeyBitmap  string   `json:"pubKeyBitmap"`
	Size          int64    `json:"size"`
	Timestamp     uint64   `json:"timestamp"`
	TxCount       uint32   `json:"txCount"`
	StateRootHash string   `json:"stateRootHash"`
	PrevHash      string   `json:"prevHash"`
}
