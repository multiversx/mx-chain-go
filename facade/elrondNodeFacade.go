package facade

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	"github.com/ElrondNetwork/elrond-go/api"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/ntp"
)

// DefaultRestPort is the default port the REST API will start on if not specified
const DefaultRestPort = "8080"

// DefaultRestPortOff is the default value that should be passed if it is desired
//  to start the node without a REST endpoint available
const DefaultRestPortOff = "off"

// ElrondNodeFacade represents a facade for grouping the functionality for node, transaction and address
type ElrondNodeFacade struct {
	node                   NodeWrapper
	apiResolver            ApiResolver
	syncer                 ntp.SyncTimer
	log                    *logger.Logger
	tpsBenchmark           *statistics.TpsBenchmark
	config                 *config.FacadeConfig
	restAPIServerDebugMode bool
}

// NewElrondNodeFacade creates a new Facade with a NodeWrapper
func NewElrondNodeFacade(node NodeWrapper, apiResolver ApiResolver, restAPIServerDebugMode bool) *ElrondNodeFacade {
	if node == nil || node.IsInterfaceNil() {
		return nil
	}
	if apiResolver == nil || apiResolver.IsInterfaceNil() {
		return nil
	}

	return &ElrondNodeFacade{
		node:                   node,
		apiResolver:            apiResolver,
		restAPIServerDebugMode: restAPIServerDebugMode,
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

// SetConfig sets the configuration options for the facade
func (ef *ElrondNodeFacade) SetConfig(facadeConfig *config.FacadeConfig) {
	ef.config = facadeConfig
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

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (ef *ElrondNodeFacade) RestAPIServerDebugMode() bool {
	return ef.restAPIServerDebugMode
}

// RestApiPort returns the port on which the api should start on, based on the config file provided.
// The API will start on the DefaultRestPort value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (ef *ElrondNodeFacade) RestApiPort() string {
	if ef.config == nil {
		return DefaultRestPort
	}
	if ef.config.RestApiPort == "" {
		return DefaultRestPort
	}
	if ef.config.RestApiPort == DefaultRestPortOff {
		return DefaultRestPortOff
	}

	_, err := strconv.ParseInt(ef.config.RestApiPort, 10, 32)
	if err != nil {
		return DefaultRestPort
	}

	return ef.config.RestApiPort
}

// PrometheusMonitoring returns if prometheus is enabled for monitoring by the flag
func (ef *ElrondNodeFacade) PrometheusMonitoring() bool {
	return ef.config.Prometheus
}

// PrometheusJoinURL will return the join URL from server.toml
func (ef *ElrondNodeFacade) PrometheusJoinURL() string {
	return ef.config.PrometheusJoinURL
}

// PrometheusNetworkID will return the NetworkID from config.toml or the flag
func (ef *ElrondNodeFacade) PrometheusNetworkID() string {
	return ef.config.PrometheusJobName
}

func (ef *ElrondNodeFacade) startRest(wg *sync.WaitGroup) {
	defer wg.Done()

	switch ef.RestApiPort() {
	case DefaultRestPortOff:
		ef.log.Info(fmt.Sprintf("Web server is off"))
		break
	default:
		ef.log.Info("Starting web server...")
		err := api.Start(ef)
		if err != nil {
			ef.log.Error("Could not start webserver", err.Error())
		}
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

// CreateTransaction creates a transaction from all needed fields
func (ef *ElrondNodeFacade) CreateTransaction(
	nonce uint64,
	value *big.Int,
	receiverHex string,
	senderHex string,
	gasPrice uint64,
	gasLimit uint64,
	data string,
	signatureHex string,
	challenge string,
) (*transaction.Transaction, error) {

	return ef.node.CreateTransaction(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, data, signatureHex, challenge)
}

// SendTransaction will send a new transaction on the topic channel
func (ef *ElrondNodeFacade) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value *big.Int,
	gasPrice uint64,
	gasLimit uint64,
	transactionData string,
	signature []byte,
) (string, error) {

	return ef.node.SendTransaction(nonce, senderHex, receiverHex, value, gasPrice, gasLimit, transactionData, signature)
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (ef *ElrondNodeFacade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return ef.node.SendBulkTransactions(txs)
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

// NodeDetails will return the node's details handler
func (ef *ElrondNodeFacade) NodeDetails() external.NodeDetailsHandler {
	return ef.apiResolver.NodeDetails()
}

// GetVmValue retrieves data from existing SC trie
func (ef *ElrondNodeFacade) GetVmValue(address string, funcName string, argsBuff ...[]byte) ([]byte, error) {
	return ef.apiResolver.GetVmValue(address, funcName, argsBuff...)
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (ef *ElrondNodeFacade) PprofEnabled() bool {
	return ef.config.PprofEnabled
}

// IsInterfaceNil returns true if there is no value under the interface
func (ef *ElrondNodeFacade) IsInterfaceNil() bool {
	if ef == nil {
		return true
	}
	return false
}
