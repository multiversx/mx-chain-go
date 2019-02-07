package mock

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

type NodeMock struct {
	AddressHandler                         func() (string, error)
	StartHandler                           func() error
	StopHandler                            func() error
	P2PBootstrapHandler                    func() error
	IsRunningHandler                       func() bool
	ConnectToAddressesHandler              func([]string) error
	StartConsensusHandler                  func() error
	GetBalanceHandler                      func(address string) (*big.Int, error)
	GenerateTransactionHandler             func(sender string, receiver string, amount *big.Int, code string) (*transaction.Transaction, error)
	GetTransactionHandler                  func(hash string) (*transaction.Transaction, error)
	SendTransactionHandler                 func(nonce uint64, sender string, receiver string, amount *big.Int, code string, signature []byte) (*transaction.Transaction, error)
	GetAccountHandler                      func(address string) (*state.Account, error)
	GetCurrentPublicKeyHandler             func() string
	GenerateAndSendBulkTransactionsHandler func(destination string, value *big.Int, nrTransactions uint64) error
}

func (nm *NodeMock) Address() (string, error) {
	return nm.AddressHandler()
}

func (nm *NodeMock) Start() error {
	return nm.StartHandler()
}

func (nm *NodeMock) Stop() error {
	return nm.StopHandler()
}

func (nm *NodeMock) P2PBootstrap() error {
	return nm.P2PBootstrapHandler()
}

func (nm *NodeMock) IsRunning() bool {
	return nm.IsRunningHandler()
}

func (nm *NodeMock) ConnectToAddresses(addresses []string) error {
	return nm.ConnectToAddressesHandler(addresses)
}

func (nm *NodeMock) StartConsensus() error {
	return nm.StartConsensusHandler()
}

func (nm *NodeMock) GetBalance(address string) (*big.Int, error) {
	return nm.GetBalanceHandler(address)
}

func (nm *NodeMock) GenerateTransaction(sender string, receiver string, amount *big.Int, code string) (*transaction.Transaction, error) {
	return nm.GenerateTransactionHandler(sender, receiver, amount, code)
}

func (nm *NodeMock) GetTransaction(hash string) (*transaction.Transaction, error) {
	return nm.GetTransactionHandler(hash)
}

func (nm *NodeMock) SendTransaction(nonce uint64, sender string, receiver string, value *big.Int, transactionData string, signature []byte) (*transaction.Transaction, error) {
	return nm.SendTransactionHandler(nonce, sender, receiver, value, transactionData, signature)
}

func (nm *NodeMock) GetInterceptors() []process.Interceptor {
	return nil
}

func (nm *NodeMock) GetResolvers() []process.Resolver {
	return nil
}

func (nm *NodeMock) GetCurrentPublicKey() string {
	return nm.GetCurrentPublicKeyHandler()
}

func (nm *NodeMock) GenerateAndSendBulkTransactions(receiverHex string, value *big.Int, noOfTx uint64) error {
	return nm.GenerateAndSendBulkTransactionsHandler(receiverHex, value, noOfTx)
}

func (nm *NodeMock) GetAccount(address string) (*state.Account, error) {
	return nm.GetAccountHandler(address)
}
