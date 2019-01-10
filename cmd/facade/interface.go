package facade

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

// Facade represents a facade for grouping the functionality needed for node, transaction and address
type Facade interface {
	//StartNode starts the underlying node
	StartNode() error

	//StopNode stops the underlying node
	StopNode() error

	//StartBackgroundServices starts all background services needed for the correct functionality of the node
	StartBackgroundServices(wg *sync.WaitGroup)

	//SetLogger sets the current logger
	SetLogger(logger *logger.Logger)

	//IsNodeRunning gets if the underlying node is running
	IsNodeRunning() bool

	//GetBalance gets the current balance for a specified address
	GetBalance(address string) (*big.Int, error)

	//GenerateTransaction generates a transaction from a sender, receiver, value and data
	GenerateTransaction(sender string, receiver string, value big.Int, data string) (*transaction.Transaction, error)

	//SendTransaction will send a new transaction on the topic channel
	SendTransaction(nonce uint64, sender string, receiver string, value big.Int, transactionData string, signature string) (*transaction.Transaction, error)

	//GetTransaction gets the transaction with a specified hash
	GetTransaction(hash string) (*transaction.Transaction, error)
}

//NodeWrapper contains all functions that a node should contain.
type NodeWrapper interface {

	// Address returns the first address of the running node
	Address() (string, error)

	// Start will create a new messenger and and set up the Node state as running
	Start() error

	// Stop closes the messenger and undos everything done in Start
	Stop() error

	// P2PBootstrap starts the peer discovery process and peer connection filtering
	P2PBootstrap()

	//IsRunning returns if the underlying node is running
	IsRunning() bool

	// ConnectToAddresses will take a slice of addresses and try to connect to all of them.
	ConnectToAddresses(addresses []string) error

	// BindInterceptorsResolvers will start the interceptors and resolvers
	BindInterceptorsResolvers() error

	// StartConsensus will start the consesus service for the current node
	StartConsensus() error

	//GetBalance returns the balance for a specific address
	GetBalance(address string) (*big.Int, error)

	//GenerateTransaction generates a new transaction with sender, receiver, amount and code
	GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error)

	//SendTransaction will send a new transaction on the topic channel
	SendTransaction(nonce uint64, sender string, receiver string, value big.Int, transactionData string, signature string) (*transaction.Transaction, error)

	//GetTransaction gets the transaction
	GetTransaction(hash string) (*transaction.Transaction, error)
}
