package facade

import (
	"errors"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"math/big"
	"strconv"
	"sync"
	"time"

	"github.com/ElrondNetwork/elrond-go-sandbox/api"
	"github.com/ElrondNetwork/elrond-go-sandbox/logger"

	"github.com/ElrondNetwork/elrond-go-sandbox/chronology/ntp"
	beevikntp "github.com/beevik/ntp"
)

type ElrondNodeFacade struct {
	node     NodeWrapper
	syncTime *ntp.SyncTime
	log      *logger.Logger
}

func NewElrondNodeFacade(node NodeWrapper) *ElrondNodeFacade {
	if node == nil {
		return nil
	}

	return &ElrondNodeFacade{
		node: node,
	}
}

func (ef *ElrondNodeFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

func (ef *ElrondNodeFacade) StartNode() error {
	err := ef.node.Start()
	if err != nil {
		return err
	}
	err = ef.node.ConnectToInitialAddresses()
	if err != nil {
		return err
	}
	err = ef.node.StartConsensus()
	return err
}

func (ef *ElrondNodeFacade) StopNode() error {
	return ef.node.Stop()
}

func (ef *ElrondNodeFacade) StartNTP(clockSyncPeriod int) {
	ef.syncTime = ntp.NewSyncTime(time.Second*time.Duration(clockSyncPeriod), func(host string) (response *beevikntp.Response, e error) {
		return nil, errors.New("this should be implemented")
	})
}

func (ef *ElrondNodeFacade) WaitForStartTime(t time.Time) {
	if !ef.syncTime.CurrentTime(ef.syncTime.ClockOffset()).After(t) {
		diff := t.Sub(ef.syncTime.CurrentTime(ef.syncTime.ClockOffset())).Seconds()
		ef.log.Info("Elrond protocol not started yet, waiting " + strconv.Itoa(int(diff)) + " seconds")
	}
	for {
		if ef.syncTime.CurrentTime(ef.syncTime.ClockOffset()).After(t) {
			break
		}
		time.Sleep(time.Duration(5 * time.Millisecond))
	}
}

func (ef *ElrondNodeFacade) StartBackgroundServices(wg *sync.WaitGroup) {
	wg.Add(1)
	go ef.startRest(wg)
}

func (ef *ElrondNodeFacade) IsNodeRunning() bool {
	return ef.node.IsRunning()
}

func (ef *ElrondNodeFacade) startRest(wg *sync.WaitGroup) {
	ef.log.Info("Starting web server...")
	err := api.Start(ef)
	if err != nil {
		ef.log.Error("Could not start webserver", err.Error())
	}
	wg.Done()
}

func (ef *ElrondNodeFacade) GetBalance(address string) (*big.Int, error) {
	return ef.node.GetBalance(address)
}

func (ef *ElrondNodeFacade) GenerateTransaction(sender string, receiver string, amount big.Int, code string) (*transaction.Transaction,
	error) {
	return ef.node.GenerateTransaction(sender, receiver, amount, code)
}

func (ef *ElrondNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}
