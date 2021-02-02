package types

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
	Fee                  string        `json:"fee"`
	Data                 []byte        `json:"data"`
	Signature            string        `json:"signature"`
	Timestamp            time.Duration `json:"timestamp"`
	Status               string        `json:"status"`
	SearchOrder          uint32        `json:"searchOrder"`
	EsdtTokenIdentifier  string        `json:"token,omitempty"`
	EsdtValue            string        `json:"esdtValue,omitempty"`
	SenderUserName       []byte        `json:"senderUserName,omitempty"`
	ReceiverUserName     []byte        `json:"receiverUserName,omitempty"`
	Logs                 *TxLog        `json:"logs,omitempty"`
	HasSCR               bool          `json:"hasScResults,omitempty"`
	SmartContractResults []*ScResult   `json:"-"`
	RcvAddrBytes         []byte        `json:"-"`
}

// GetGasLimit will return transaction gas limit
func (t *Transaction) GetGasLimit() uint64 {
	return t.GasLimit
}

// GetGasPrice will return transaction gas price
func (t *Transaction) GetGasPrice() uint64 {
	return t.GasPrice
}

// GetData will return transaction data field
func (t *Transaction) GetData() []byte {
	return t.Data
}

// GetRcvAddr will return transaction receiver address
func (t *Transaction) GetRcvAddr() []byte {
	return t.RcvAddrBytes
}

// GetValue wil return transaction value
func (t *Transaction) GetValue() *big.Int {
	bigIntValue, ok := big.NewInt(0).SetString(t.Value, 10)
	if !ok {
		return big.NewInt(0)
	}

	return bigIntValue
}

// Receipt is a structure containing all the fields that need to be save for a Receipt
type Receipt struct {
	Hash      string        `json:"-"`
	Value     string        `json:"value"`
	Sender    string        `json:"sender"`
	Data      string        `json:"data,omitempty"`
	TxHash    string        `json:"txHash"`
	Timestamp time.Duration `json:"timestamp"`
}

// ScResult is a structure containing all the fields that need to be saved for a smart contract result
type ScResult struct {
	Hash                string        `json:"-"`
	Nonce               uint64        `json:"nonce"`
	GasLimit            uint64        `json:"gasLimit"`
	GasPrice            uint64        `json:"gasPrice"`
	Value               string        `json:"value"`
	Sender              string        `json:"sender"`
	Receiver            string        `json:"receiver"`
	RelayerAddr         string        `json:"relayerAddr,omitempty"`
	RelayedValue        string        `json:"relayedValue,omitempty"`
	Code                string        `json:"code,omitempty"`
	Data                []byte        `json:"data,omitempty"`
	PreTxHash           string        `json:"prevTxHash"`
	OriginalTxHash      string        `json:"originalTxHash"`
	CallType            string        `json:"callType"`
	CodeMetadata        []byte        `json:"codeMetaData,omitempty"`
	ReturnMessage       string        `json:"returnMessage,omitempty"`
	Timestamp           time.Duration `json:"timestamp"`
	EsdtTokenIdentifier string        `json:"token,omitempty"`
	EsdtValue           string        `json:"esdtValue,omitempty"`
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

// PreparedResults is the TDO that holds all the results after processing
type PreparedResults struct {
	Transactions    []*Transaction
	ScResults       []*ScResult
	Receipts        []*Receipt
	AlteredAccounts map[string]*AlteredAccount
}
