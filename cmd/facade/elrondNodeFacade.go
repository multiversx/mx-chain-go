package facade

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/api"
	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"
)

// ElrondNodeFacade represents a facade for grouping the functionality for node, transaction and address
type ElrondNodeFacade struct {
	node   NodeWrapper
	syncer ntp.SyncTimer
	log    *logger.Logger
}

// NewElrondNodeFacade creates a new Facade with a NodeWrapper
func NewElrondNodeFacade(node NodeWrapper) *ElrondNodeFacade {
	if node == nil {
		return nil
	}

	return &ElrondNodeFacade{
		node: node,
	}
}

// SetLogger sets the current logger
func (ef *ElrondNodeFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

//SetSyncer sets the current syncer
func (ef *ElrondNodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	ef.syncer = syncer
}

// StartNode starts the underlying node
func (ef *ElrondNodeFacade) StartNode() error {
	err := ef.node.Start()
	if err != nil {
		return err
	}

	err = ef.node.StartConsensus()
	return err
}

// StopNode stops the underlying node
func (ef *ElrondNodeFacade) StopNode() error {
	return ef.node.Stop()
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (ef *ElrondNodeFacade) StartBackgroundServices(wg *sync.WaitGroup) {
	wg.Add(1)
	go ef.startRest(wg)
}

// IsNodeRunning gets if the underlying node is running
func (ef *ElrondNodeFacade) IsNodeRunning() bool {
	return ef.node.IsRunning()
}

func (ef *ElrondNodeFacade) startRest(wg *sync.WaitGroup) {
	defer wg.Done()

	ef.log.Info("Starting web server...")
	err := api.Start(ef)
	if err != nil {
		ef.log.Error("Could not start webserver", err.Error())
	}
}

// GetBalance gets the current balance for a specified address
func (ef *ElrondNodeFacade) GetBalance(address string) (*big.Int, error) {
	return ef.node.GetBalance(address)
}

// GenerateTransaction generates a transaction from a sender, receiver, value and data
func (ef *ElrondNodeFacade) GenerateTransaction(sender string, receiver string, value big.Int,
	data string) (*transaction.Transaction,
	error) {
	return ef.node.GenerateTransaction(sender, receiver, value, data)
}

// SendTransaction will send a new transaction on the topic channel
func (ef *ElrondNodeFacade) SendTransaction(nonce uint64, sender string, receiver string,
	value big.Int, transactionData string, signature string) (*transaction.Transaction, error) {
	return ef.node.SendTransaction(nonce, sender, receiver, value, transactionData, signature)
}

// GetTransaction gets the transaction with a specified hash
func (ef *ElrondNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}
