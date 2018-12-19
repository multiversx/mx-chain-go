package facade

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

//NodeWrapper contains all functions that a node should contain.
type NodeWrapper interface {

	// Address returns the first address of the running node
	Address() (string, error)

	// Start will create a new messenger and and set up the Node state as running
	Start() error

	// Stop closes the messenger and undos everything done in Start
	Stop() error

	//Returns if the underlying node is running
	IsRunning() bool

	// ConnectToInitialAddresses connect to the list of peers provided initialAddresses
	ConnectToInitialAddresses() error

	// ConnectToAddresses will take a slice of addresses and try to connect to all of them.
	ConnectToAddresses(addresses []string) error

	// StartConsensus will start the consesus service for the current node
	StartConsensus() error

	//GetBalance returns the balance for a specific address
	GetBalance(address string) (*big.Int, error)

	//GenerateTransaction generates a new transaction with sender, receiver, amount and code
	GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction, error)

	//GetTransaction gets the transaction
	GetTransaction(hash string) (*transaction.Transaction, error)
}
