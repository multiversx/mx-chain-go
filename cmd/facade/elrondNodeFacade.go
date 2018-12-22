package facade

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/api"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
)

//Facade for grouping the functionality for node, transaction and address
type ElrondNodeFacade struct {
	node   NodeWrapper
	syncer ntp.SyncTimer
	log    *logger.Logger
}

//Creates a new Facade with a NodeWrapper
func NewElrondNodeFacade(node NodeWrapper) *ElrondNodeFacade {
	if node == nil {
		return nil
	}

	return &ElrondNodeFacade{
		node: node,
	}
}

//Sets the current logger
func (ef *ElrondNodeFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

//Sets the current syncer
func (ef *ElrondNodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	ef.syncer = syncer
}

//Starts the underlying node
func (ef *ElrondNodeFacade) StartNode() error {
	err := ef.node.Start()
	if err != nil {
		return err
	}
	//err = ef.node.ConnectToInitialAddresses()
	//if err != nil {
	//	return err
	//}
	err = ef.node.StartConsensus()
	return err
}

//Stops the underlying node
func (ef *ElrondNodeFacade) StopNode() error {
	return ef.node.Stop()
}

//Waits for the startTime to arrive and only after proceeds
func (ef *ElrondNodeFacade) WaitForStartTime(t time.Time) {
	if !ef.syncer.CurrentTime(ef.syncer.ClockOffset()).After(t) {
		diff := t.Sub(ef.syncer.CurrentTime(ef.syncer.ClockOffset())).Seconds()
		ef.log.Info("Elrond protocol not started yet, waiting " + strconv.Itoa(int(diff)) + " seconds")
	}
	for {
		if ef.syncer.CurrentTime(ef.syncer.ClockOffset()).After(t) {
			break
		}
		time.Sleep(time.Duration(5 * time.Millisecond))
	}
}

//Starts all background services needed for the correct functionality of the node
func (ef *ElrondNodeFacade) StartBackgroundServices(wg *sync.WaitGroup) {
	wg.Add(1)
	go ef.startRest(wg)
}

//Gets if the underlying node is running
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

//Gets the current balance for a specified address
func (ef *ElrondNodeFacade) GetBalance(address string) (*big.Int, error) {
	return ef.node.GetBalance(address)
}

//Generates a transaction from a sender, receiver, value and data
func (ef *ElrondNodeFacade) GenerateTransaction(sender string, receiver string, value big.Int,
	data string) (*transaction.Transaction,
	error) {
	return ef.node.GenerateTransaction(sender, receiver, value, data)
}

//Gets the transaction with a specified hash
func (ef *ElrondNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}
