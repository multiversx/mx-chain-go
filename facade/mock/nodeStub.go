package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
)

// NodeStub -
type NodeStub struct {
	AddressHandler             func() (string, error)
	StartHandler               func() error
	StopHandler                func() error
	P2PBootstrapHandler        func() error
	IsRunningHandler           func() bool
	ConnectToAddressesHandler  func([]string) error
	StartConsensusHandler      func() error
	GetBalanceHandler          func(address string) (*big.Int, error)
	GenerateTransactionHandler func(sender string, receiver string, amount string, code string) (*transaction.Transaction, error)
	CreateTransactionHandler   func(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64,
		gasLimit uint64, data []byte, signatureHex string) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler                     func(tx *transaction.Transaction) error
	GetTransactionHandler                          func(hash string) (*transaction.Transaction, error)
	SendBulkTransactionsHandler                    func(txs []*transaction.Transaction) (uint64, error)
	GetAccountHandler                              func(address string) (state.UserAccountHandler, error)
	GetCurrentPublicKeyHandler                     func() string
	GenerateAndSendBulkTransactionsHandler         func(destination string, value *big.Int, nrTransactions uint64) error
	GenerateAndSendBulkTransactionsOneByOneHandler func(destination string, value *big.Int, nrTransactions uint64) error
	GetHeartbeatsHandler                           func() []heartbeat.PubKeyHeartbeat
	ValidatorStatisticsApiCalled                   func() (map[string]*state.ValidatorApiResponse, error)
	DirectTriggerCalled                            func() error
	IsSelfTriggerCalled                            func() bool
}

// Address -
func (ns *NodeStub) Address() (string, error) {
	return ns.AddressHandler()
}

// Start -
func (ns *NodeStub) Start() error {
	return ns.StartHandler()
}

// P2PBootstrap -
func (ns *NodeStub) P2PBootstrap() error {
	return ns.P2PBootstrapHandler()
}

// IsRunning -
func (ns *NodeStub) IsRunning() bool {
	return ns.IsRunningHandler()
}

// ConnectToAddresses -
func (ns *NodeStub) ConnectToAddresses(addresses []string) error {
	return ns.ConnectToAddressesHandler(addresses)
}

// StartConsensus -
func (ns *NodeStub) StartConsensus() error {
	return ns.StartConsensusHandler()
}

// GetBalance -
func (ns *NodeStub) GetBalance(address string) (*big.Int, error) {
	return ns.GetBalanceHandler(address)
}

// GenerateTransaction -
func (ns *NodeStub) GenerateTransaction(sender string, receiver string, amount string, code string) (*transaction.Transaction, error) {
	return ns.GenerateTransactionHandler(sender, receiver, amount, code)
}

// CreateTransaction -
func (ns *NodeStub) CreateTransaction(nonce uint64, value string, receiverHex string, senderHex string, gasPrice uint64,
	gasLimit uint64, data []byte, signatureHex string) (*transaction.Transaction, []byte, error) {

	return ns.CreateTransactionHandler(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, data, signatureHex)
}

//ValidateTransaction --
func (ns *NodeStub) ValidateTransaction(tx *transaction.Transaction) error {
	return ns.ValidateTransactionHandler(tx)
}

// GetTransaction -
func (ns *NodeStub) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ns.GetTransactionHandler(hash)
}

// SendBulkTransactions -
func (ns *NodeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return ns.SendBulkTransactionsHandler(txs)
}

// GetCurrentPublicKey -
func (ns *NodeStub) GetCurrentPublicKey() string {
	return ns.GetCurrentPublicKeyHandler()
}

// GenerateAndSendBulkTransactions -
func (ns *NodeStub) GenerateAndSendBulkTransactions(receiverHex string, value *big.Int, noOfTx uint64) error {
	return ns.GenerateAndSendBulkTransactionsHandler(receiverHex, value, noOfTx)
}

// GenerateAndSendBulkTransactionsOneByOne -
func (ns *NodeStub) GenerateAndSendBulkTransactionsOneByOne(receiverHex string, value *big.Int, noOfTx uint64) error {
	return ns.GenerateAndSendBulkTransactionsOneByOneHandler(receiverHex, value, noOfTx)
}

// GetAccount -
func (ns *NodeStub) GetAccount(address string) (*state.Account, error) {
	return ns.GetAccountHandler(address)
func (nm *NodeMock) GetAccount(address string) (state.UserAccountHandler, error) {
	return nm.GetAccountHandler(address)
}

// GetHeartbeats -
func (ns *NodeStub) GetHeartbeats() []heartbeat.PubKeyHeartbeat {
	return ns.GetHeartbeatsHandler()
}

// ValidatorStatisticsApi -
func (ns *NodeStub) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return ns.ValidatorStatisticsApiCalled()
}

// DirectTrigger -
func (ns *NodeStub) DirectTrigger() error {
	return ns.DirectTriggerCalled()
}

// IsSelfTrigger -
func (ns *NodeStub) IsSelfTrigger() bool {
	return ns.IsSelfTriggerCalled()
}

// IsInterfaceNil returns true if there is no value under the interface
func (ns *NodeStub) IsInterfaceNil() bool {
	return ns == nil
}
