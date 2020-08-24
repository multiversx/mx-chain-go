package transaction

import (
	"math/big"

	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ApiTransactionResult is the data transfer object which will be returned on the get transaction by hash endpoint
type ApiTransactionResult struct {
	Type                              string   `json:"type"`
	Hash                              string   `json:"hash,omitempty"`
	Nonce                             uint64   `json:"nonce,omitempty"`
	Round                             uint64   `json:"round,omitempty"`
	Epoch                             uint32   `json:"epoch,omitempty"`
	Value                             string   `json:"value,omitempty"`
	Receiver                          string   `json:"receiver,omitempty"`
	Sender                            string   `json:"sender,omitempty"`
	GasPrice                          uint64   `json:"gasPrice,omitempty"`
	GasLimit                          uint64   `json:"gasLimit,omitempty"`
	Data                              []byte   `json:"data,omitempty"`
	Code                              string   `json:"code,omitempty"`
	Signature                         string   `json:"signature,omitempty"`
	SourceShard                       uint32   `json:"sourceShard"`
	DestinationShard                  uint32   `json:"destinationShard"`
	BlockNonce                        uint64   `json:"blockNonce,omitempty"`
	BlockHash                         string   `json:"blockHash,omitempty"`
	NotarizedAtSourceInMetaNonce      uint64   `json:"notarizedAtSourceInMetaNonce,omitempty"`
	NotarizedAtSourceInMetaHash       string   `json:"NotarizedAtSourceInMetaHash,omitempty"`
	NotarizedAtDestinationInMetaNonce uint64   `json:"notarizedAtDestinationInMetaNonce,omitempty"`
	NotarizedAtDestinationInMetaHash  string   `json:"notarizedAtDestinationInMetaHash,omitempty"`
	MiniBlockHash                     string   `json:"miniblockHash,omitempty"`
	Status                            TxStatus `json:"status,omitempty"`
}

// SimulationResults is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResults struct {
	Status     TxStatus                           `json:"status,omitempty"`
	FailReason string                             `json:"failReason,omitempty"`
	ScResults  map[string]*SmartContractResultApi `json:"scResults,omitempty"`
	Receipts   map[string]*ReceiptApi             `json:"receipts,omitempty"`
	Hash       string                             `json:"hash,omitempty"`
}

// SmartContractResultApi represents a smart contract result with changed fields' types in order to make it friendly for API's json
type SmartContractResultApi struct {
	Nonce          uint64            `json:"nonce"`
	Value          *big.Int          `json:"value"`
	RcvAddr        string            `json:"receiver"`
	SndAddr        string            `json:"sender"`
	RelayerAddr    string            `json:"relayerAddress"`
	RelayedValue   *big.Int          `json:"relayedValue"`
	Code           string            `json:"code"`
	Data           string            `json:"data"`
	PrevTxHash     string            `json:"prevTxHash"`
	OriginalTxHash string            `json:"originalTxHash"`
	GasLimit       uint64            `json:"gasLimit"`
	GasPrice       uint64            `json:"gasPrice"`
	CallType       vmcommon.CallType `json:"callType"`
	CodeMetadata   string            `json:"codeMetadata"`
	ReturnMessage  string            `json:"returnMessage"`
	OriginalSender string            `json:"originalSender"`
}

// ReceiptApi represents a receipt with changed fields' types in order to make it friendly for API's json
type ReceiptApi struct {
	Value   *big.Int `json:"value"`
	SndAddr string   `json:"sender"`
	Data    string   `json:"data,omitempty"`
	TxHash  string   `json:"txHash"`
}
