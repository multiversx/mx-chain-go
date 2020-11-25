package transaction

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data"
)

// ApiTransactionResult is the data transfer object which will be returned on the get transaction by hash endpoint
type ApiTransactionResult struct {
	Tx                                data.TransactionHandler   `json:"-"`
	Type                              string                    `json:"type"`
	Hash                              string                    `json:"hash,omitempty"`
	Nonce                             uint64                    `json:"nonce,omitempty"`
	Round                             uint64                    `json:"round,omitempty"`
	Epoch                             uint32                    `json:"epoch,omitempty"`
	Value                             string                    `json:"value,omitempty"`
	Receiver                          string                    `json:"receiver,omitempty"`
	Sender                            string                    `json:"sender,omitempty"`
	GasPrice                          uint64                    `json:"gasPrice,omitempty"`
	GasLimit                          uint64                    `json:"gasLimit,omitempty"`
	Data                              []byte                    `json:"data,omitempty"`
	CodeMetadata                      []byte                    `json:"codeMetadata,omitempty"`
	Code                              string                    `json:"code,omitempty"`
	PreviousTransactionHash           string                    `json:"previousTransactionHash,omitempty"`
	OriginalTransactionHash           string                    `json:"originalTransactionHash,omitempty"`
	ReturnMessage                     string                    `json:"returnMessage,omitempty"`
	OriginalSender                    string                    `json:"originalSender,omitempty"`
	Signature                         string                    `json:"signature,omitempty"`
	SourceShard                       uint32                    `json:"sourceShard"`
	DestinationShard                  uint32                    `json:"destinationShard"`
	BlockNonce                        uint64                    `json:"blockNonce,omitempty"`
	BlockHash                         string                    `json:"blockHash,omitempty"`
	NotarizedAtSourceInMetaNonce      uint64                    `json:"notarizedAtSourceInMetaNonce,omitempty"`
	NotarizedAtSourceInMetaHash       string                    `json:"NotarizedAtSourceInMetaHash,omitempty"`
	NotarizedAtDestinationInMetaNonce uint64                    `json:"notarizedAtDestinationInMetaNonce,omitempty"`
	NotarizedAtDestinationInMetaHash  string                    `json:"notarizedAtDestinationInMetaHash,omitempty"`
	MiniBlockType                     string                    `json:"miniblockType,omitempty"`
	MiniBlockHash                     string                    `json:"miniblockHash,omitempty"`
	Receipt                           *ReceiptApi               `json:"receipt,omitempty"`
	SmartContractResults              []*ApiSmartContractResult `json:"smartContractResults,omitempty"`
	Status                            TxStatus                  `json:"status,omitempty"`
}

// SimulationResults is the data transfer object which will hold results for simulation a transaction's execution
type SimulationResults struct {
	Status     TxStatus                           `json:"status,omitempty"`
	FailReason string                             `json:"failReason,omitempty"`
	ScResults  map[string]*ApiSmartContractResult `json:"scResults,omitempty"`
	Receipts   map[string]*ReceiptApi             `json:"receipts,omitempty"`
	Hash       string                             `json:"hash,omitempty"`
}

// ApiSmartContractResult represents a smart contract result with changed fields' types in order to make it friendly for API's json
type ApiSmartContractResult struct {
	Hash           string            `json:"hash,omitempty"`
	Nonce          uint64            `json:"nonce"`
	Value          *big.Int          `json:"value"`
	RcvAddr        string            `json:"receiver"`
	SndAddr        string            `json:"sender"`
	RelayerAddr    string            `json:"relayerAddress,omitempty"`
	RelayedValue   *big.Int          `json:"relayedValue,omitempty"`
	Code           string            `json:"code,omitempty"`
	Data           string            `json:"data,omitempty"`
	PrevTxHash     string            `json:"prevTxHash"`
	OriginalTxHash string            `json:"originalTxHash"`
	GasLimit       uint64            `json:"gasLimit"`
	GasPrice       uint64            `json:"gasPrice"`
	CallType       vmcommon.CallType `json:"callType"`
	CodeMetadata   string            `json:"codeMetadata,omitempty"`
	ReturnMessage  string            `json:"returnMessage,omitempty"`
	OriginalSender string            `json:"originalSender,omitempty"`
}

// ReceiptApi represents a receipt with changed fields' types in order to make it friendly for API's json
type ReceiptApi struct {
	Value   *big.Int `json:"value"`
	SndAddr string   `json:"sender"`
	Data    string   `json:"data,omitempty"`
	TxHash  string   `json:"txHash"`
}
