package mock

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// Facade is the mock implementation of a node router handler
type Facade struct {
	Running                                        bool
	ShouldErrorStart                               bool
	ShouldErrorStop                                bool
	GetCurrentPublicKeyHandler                     func() string
	TpsBenchmarkHandler                            func() *statistics.TpsBenchmark
	GetHeartbeatsHandler                           func() ([]heartbeat.PubKeyHeartbeat, error)
	BalanceHandler                                 func(string) (*big.Int, error)
	GetAccountHandler                              func(address string) (*state.Account, error)
	GenerateTransactionHandler                     func(sender string, receiver string, value *big.Int, code string) (*transaction.Transaction, error)
	GetTransactionHandler                          func(hash string) (*transaction.Transaction, error)
	SendTransactionHandler                         func(nonce uint64, sender string, receiver string, value string, gasPrice uint64, gasLimit uint64, code string, signature []byte) (string, error)
	CreateTransactionHandler                       func(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64, gasLimit uint64, data string, signatureHex string, challenge string) (*transaction.Transaction, error)
	SendBulkTransactionsHandler                    func(txs []*transaction.Transaction) (uint64, error)
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
	ExecuteSCQueryHandler                          func(query *smartContract.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler                           func() external.StatusMetricsHandler
}

// IsNodeRunning is the mock implementation of a handler's IsNodeRunning method
func (f *Facade) IsNodeRunning() bool {
	return f.Running
}

// StartNode is the mock implementation of a handler's StartNode method
func (f *Facade) StartNode() error {
	if f.ShouldErrorStart {
		return errors.New("error")
	}
	return nil
}

// TpsBenchmark is the mock implementation for retreiving the TpsBenchmark
func (f *Facade) TpsBenchmark() *statistics.TpsBenchmark {
	if f.TpsBenchmarkHandler != nil {
		return f.TpsBenchmarkHandler()
	}
	return nil
}

// StopNode is the mock implementation of a handler's StopNode method
func (f *Facade) StopNode() error {
	if f.ShouldErrorStop {
		return errors.New("error")
	}
	f.Running = false
	return nil
}

// GetCurrentPublicKey is the mock implementation of a handler's StopNode method
func (f *Facade) GetCurrentPublicKey() string {
	return f.GetCurrentPublicKeyHandler()
}

func (f *Facade) GetHeartbeats() ([]heartbeat.PubKeyHeartbeat, error) {
	return f.GetHeartbeatsHandler()
}

// GetBalance is the mock implementation of a handler's GetBalance method
func (f *Facade) GetBalance(address string) (*big.Int, error) {
	return f.BalanceHandler(address)
}

// GetAccount is the mock implementation of a handler's GetAccount method
func (f *Facade) GetAccount(address string) (*state.Account, error) {
	return f.GetAccountHandler(address)
}

// GenerateTransaction is the mock implementation of a handler's GenerateTransaction method
func (f *Facade) GenerateTransaction(sender string, receiver string, value *big.Int,
	code string) (*transaction.Transaction, error) {
	return f.GenerateTransactionHandler(sender, receiver, value, code)
}

// CreateTransaction is  mock implementation of a handler's CreateTransaction method
func (f *Facade) CreateTransaction(
	nonce uint64,
	value string,
	receiverHex string,
	senderHex string,
	gasPrice uint64,
	gasLimit uint64,
	data string,
	signatureHex string,
	challenge string,
) (*transaction.Transaction, error) {

	return f.CreateTransactionHandler(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, data, signatureHex, challenge)
}

// GetTransaction is the mock implementation of a handler's GetTransaction method
func (f *Facade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return f.GetTransactionHandler(hash)
}

// SendTransaction is the mock implementation of a handler's SendTransaction method
func (f *Facade) SendTransaction(nonce uint64, sender string, receiver string, value string, gasPrice uint64, gasLimit uint64, code string, signature []byte) (string, error) {
	return f.SendTransactionHandler(nonce, sender, receiver, value, gasPrice, gasLimit, code, signature)
}

// SendBulkTransactions is the mock implementation of a handler's SendBulkTransactions method
func (f *Facade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return f.SendBulkTransactionsHandler(txs)
}

// GenerateAndSendBulkTransactions is the mock implementation of a handler's GenerateAndSendBulkTransactions method
func (f *Facade) GenerateAndSendBulkTransactions(destination string, value *big.Int, nrTransactions uint64) error {
	return f.GenerateAndSendBulkTransactionsHandler(destination, value, nrTransactions)
}

// GenerateAndSendBulkTransactionsOneByOne is the mock implementation of a handler's GenerateAndSendBulkTransactionsOneByOne method
func (f *Facade) GenerateAndSendBulkTransactionsOneByOne(destination string, value *big.Int, nrTransactions uint64) error {
	return f.GenerateAndSendBulkTransactionsOneByOneHandler(destination, value, nrTransactions)
}

// ExecuteSCQuery is a mock implementation.
func (f *Facade) ExecuteSCQuery(query *smartContract.SCQuery) (*vmcommon.VMOutput, error) {
	return f.ExecuteSCQueryHandler(query)
}

// StatusMetrics is the mock implementation for the StatusMetrics
func (f *Facade) StatusMetrics() external.StatusMetricsHandler {
	return f.StatusMetricsHandler()
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *Facade) IsInterfaceNil() bool {
	if f == nil {
		return true
	}
	return false
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
