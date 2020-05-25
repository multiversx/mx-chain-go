package indexer

import (
	"math/big"
	"time"
)

// Transaction is a structure containing all the fields that need
//  to be saved for a transaction. It has all the default fields
//  plus some extra information for ease of search and filter
type Transaction struct {
	Hash                 string        `json:"-"`
	MBHash               string        `json:"miniBlockHash"`
	BlockHash            string        `json:"-"`
	Nonce                uint64        `json:"nonce"`
	Round                uint64        `json:"round"`
	Value                string        `json:"value"`
	Receiver             string        `json:"receiver"`
	Sender               string        `json:"sender"`
	ReceiverShard        uint32        `json:"receiverShard"`
	SenderShard          uint32        `json:"senderShard"`
	GasPrice             uint64        `json:"gasPrice"`
	GasLimit             uint64        `json:"gasLimit"`
	GasUsed              uint64        `json:"gasUsed"`
	Data                 string        `json:"data"`
	Signature            string        `json:"signature"`
	Timestamp            time.Duration `json:"timestamp"`
	Status               string        `json:"status"`
	SmartContractResults []ScResult    `json:"scResults"`
	Log                  TxLog         `json:"-"`
}

// TxLog holds all the data needed for a log structure
type TxLog struct {
	Address string  `json:"scAddress"`
	Events  []Event `json:"events"`
}

// Event holds all the data needed for an event structure
type Event struct {
	Address    string   `json:"address"`
	Identifier string   `json:"identifier"`
	Topics     []string `json:"topics"`
	Data       string   `json:"data"`
}

// ScResult is a structure containing all the fields that need to be saved for a smart contract result
type ScResult struct {
	Nonce         uint64 `json:"nonce"`
	GasLimit      uint64 `json:"gasLimit"`
	GasPrice      uint64 `json:"gasPrice"`
	Value         string `json:"value"`
	Sender        string `json:"sender"`
	Receiver      string `json:"receiver"`
	Code          string `json:"code"`
	Data          string `json:"data"`
	PreTxHash     string `json:"prevTxHash"`
	CallType      string `json:"callType"`
	CodeMetadata  string `json:"codeMetaData"`
	ReturnMessage string `json:"returnMessage"`
}

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
}

//ValidatorsPublicKeys is a structure containing fields for validators public keys
type ValidatorsPublicKeys struct {
	PublicKeys []string `json:"publicKeys"`
}

// RoundInfo is a structure containing block signers and shard id
type RoundInfo struct {
	Index            uint64        `json:"round"`
	SignersIndexes   []uint64      `json:"signersIndexes"`
	BlockWasProposed bool          `json:"blockWasProposed"`
	ShardId          uint32        `json:"shardId"`
	Timestamp        time.Duration `json:"timestamp"`
}

// ValidatorsRatingInfo is a structure containing validators information
type ValidatorsRatingInfo struct {
	ValidatorsInfos []ValidatorRatingInfo `json:"validatorsRating"`
}

// ValidatorRatingInfo is a structure containing validator rating information
type ValidatorRatingInfo struct {
	PublicKey string  `json:"publicKey"`
	Rating    float32 `json:"rating"`
}

// Miniblock is a structure containing miniblock information
type Miniblock struct {
	Hash              string `json:"-"`
	SenderShardID     uint32 `json:"senderShard"`
	ReceiverShardID   uint32 `json:"receiverShard"`
	SenderBlockHash   string `json:"senderBlockHash"`
	ReceiverBlockHash string `json:"receiverBlockHash"`
	Type              string `json:"type"`
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
