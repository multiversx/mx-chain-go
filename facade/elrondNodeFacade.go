package facade

import (
	"math/big"
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/api"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/logger"
	"github.com/ElrondNetwork/elrond-go-sandbox/core/statistics"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/external"
	"github.com/ElrondNetwork/elrond-go-sandbox/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
)

// ElrondNodeFacade represents a facade for grouping the functionality for node, transaction and address
type ElrondNodeFacade struct {
	node         NodeWrapper
	resolver     ExternalResolver
	syncer       ntp.SyncTimer
	log          *logger.Logger
	tpsBenchmark *statistics.TpsBenchmark
}

// NewElrondNodeFacade creates a new Facade with a NodeWrapper
func NewElrondNodeFacade(node NodeWrapper, resolver ExternalResolver) *ElrondNodeFacade {
	if node == nil {
		return nil
	}
	if resolver == nil {
		return nil
	}

	return &ElrondNodeFacade{
		node:     node,
		resolver: resolver,
	}
}

// SetLogger sets the current logger
func (ef *ElrondNodeFacade) SetLogger(log *logger.Logger) {
	ef.log = log
}

// SetSyncer sets the current syncer
func (ef *ElrondNodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	ef.syncer = syncer
}

// SetTpsBenchmark sets the tps benchmark handler
func (ef *ElrondNodeFacade) SetTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark) {
	ef.tpsBenchmark = tpsBenchmark
}

// TpsBenchmark returns the tps benchmark handler
func (ef *ElrondNodeFacade) TpsBenchmark() *statistics.TpsBenchmark {
	return ef.tpsBenchmark
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
func (ef *ElrondNodeFacade) GenerateTransaction(senderHex string, receiverHex string, value *big.Int,
	data string) (*transaction.Transaction,
	error) {
	return ef.node.GenerateTransaction(senderHex, receiverHex, value, data)
}

// SendTransaction will send a new transaction on the topic channel
func (ef *ElrondNodeFacade) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value *big.Int,
	transactionData string,
	signature []byte,
) (*transaction.Transaction, error) {

	return ef.node.SendTransaction(nonce, senderHex, receiverHex, value, transactionData, signature)
}

// GetTransaction gets the transaction with a specified hash
func (ef *ElrondNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (ef *ElrondNodeFacade) GetAccount(address string) (*state.Account, error) {
	return ef.node.GetAccount(address)
}

// GetCurrentPublicKey gets the current nodes public Key
func (ef *ElrondNodeFacade) GetCurrentPublicKey() string {
	return ef.node.GetCurrentPublicKey()
}

// GenerateAndSendBulkTransactions generates a number of nrTransactions of amount value
// for the receiver destination
func (ef *ElrondNodeFacade) GenerateAndSendBulkTransactions(
	destination string,
	value *big.Int,
	nrTransactions uint64,
) error {

	return ef.node.GenerateAndSendBulkTransactions(destination, value, nrTransactions)
}

// GenerateAndSendBulkTransactionsOneByOne generates a number of nrTransactions of amount value
// for the receiver destination in a one by one fashion
func (ef *ElrondNodeFacade) GenerateAndSendBulkTransactionsOneByOne(
	destination string,
	value *big.Int,
	nrTransactions uint64,
) error {

	return ef.node.GenerateAndSendBulkTransactionsOneByOne(destination, value, nrTransactions)
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (ef *ElrondNodeFacade) GetHeartbeats() ([]heartbeat.PubKeyHeartbeat, error) {
	hbStatus := ef.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// RecentNotarizedBlocks computes last notarized [maxShardHeadersNum] shard headers (by metachain node)
func (ef *ElrondNodeFacade) RecentNotarizedBlocks(maxShardHeadersNum int) ([]*external.BlockHeader, error) {
	return ef.resolver.RecentNotarizedBlocks(maxShardHeadersNum)
}

// RetrieveShardBlock retrieves a shard block info containing header and transactions
func (ef *ElrondNodeFacade) RetrieveShardBlock(blockHash []byte) (*external.ShardBlockInfo, error) {
	return ef.resolver.RetrieveShardBlock(blockHash)
}
