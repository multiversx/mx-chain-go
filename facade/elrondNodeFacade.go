package facade

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/api"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// DefaultRestInterface is the default interface the rest API will start on if not specified
const DefaultRestInterface = "localhost:8080"

// DefaultRestPortOff is the default value that should be passed if it is desired
//  to start the node without a REST endpoint available
const DefaultRestPortOff = "off"

var log = logger.GetOrCreate("facade")

// ElrondNodeFacade represents a facade for grouping the functionality for node, transaction and address
type ElrondNodeFacade struct {
	node                   NodeWrapper
	apiResolver            ApiResolver
	syncer                 ntp.SyncTimer
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
func (ef *ElrondNodeFacade) StartNode(epoch uint32) error {
	err := ef.node.Start()
	if err != nil {
		return err
	}

	err = ef.node.StartConsensus(epoch)
	return err
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (ef *ElrondNodeFacade) StartBackgroundServices() {
	go ef.startRest()
}

// IsNodeRunning gets if the underlying node is running
func (ef *ElrondNodeFacade) IsNodeRunning() bool {
	return ef.node.IsRunning()
}

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (ef *ElrondNodeFacade) RestAPIServerDebugMode() bool {
	return ef.restAPIServerDebugMode
}

// RestApiInterface returns the interface on which the rest API should start on, based on the config file provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (ef *ElrondNodeFacade) RestApiInterface() string {
	if ef.config == nil {
		return DefaultRestInterface
	}
	if ef.config.RestApiInterface == "" {
		return DefaultRestInterface
	}

	return ef.config.RestApiInterface
}

func (ef *ElrondNodeFacade) startRest() {
	log.Trace("starting REST api server")

	switch ef.RestApiInterface() {
	case DefaultRestPortOff:
		log.Debug("web server is off")
	default:
		log.Debug("starting web server")
		err := api.Start(ef)
		if err != nil {
			log.Debug("could not start webserver",
				"error", err.Error(),
			)
		}
	}
}

// GetBalance gets the current balance for a specified address
func (ef *ElrondNodeFacade) GetBalance(address string) (*big.Int, error) {
	return ef.node.GetBalance(address)
}

// CreateTransaction creates a transaction from all needed fields
func (ef *ElrondNodeFacade) CreateTransaction(
	nonce uint64,
	value string,
	receiverHex string,
	senderHex string,
	gasPrice uint64,
	gasLimit uint64,
	txData []byte,
	signatureHex string,
) (*transaction.Transaction, error) {

	return ef.node.CreateTransaction(nonce, value, receiverHex, senderHex, gasPrice, gasLimit, txData, signatureHex)
}

// ValidatorStatisticsApi will return the statistics for all validators
func (ef *ElrondNodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return ef.node.ValidatorStatisticsApi()
}

// SendTransaction will send a new transaction on the topic channel
func (ef *ElrondNodeFacade) SendTransaction(
	nonce uint64,
	senderHex string,
	receiverHex string,
	value string,
	gasPrice uint64,
	gasLimit uint64,
	txData []byte,
	signature []byte,
) (string, error) {

	return ef.node.SendTransaction(nonce, senderHex, receiverHex, value, gasPrice, gasLimit, txData, signature)
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (ef *ElrondNodeFacade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return ef.node.SendBulkTransactions(txs)
}

// GetTransaction gets the transaction with a specified hash
func (ef *ElrondNodeFacade) GetTransaction(hash string) (*transaction.Transaction, error) {
	return ef.node.GetTransaction(hash)
}

// ComputeTransactionCost will estimate how many gas a transaction will consume
func (ef *ElrondNodeFacade) ComputeTransactionCost(tx *transaction.Transaction) (*big.Int, error) {
	return ef.apiResolver.ComputeTransactionCost(tx)
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (ef *ElrondNodeFacade) GetAccount(address string) (*state.Account, error) {
	return ef.node.GetAccount(address)
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (ef *ElrondNodeFacade) GetHeartbeats() ([]heartbeat.PubKeyHeartbeat, error) {
	hbStatus := ef.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// StatusMetrics will return the node's status metrics
func (ef *ElrondNodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return ef.apiResolver.StatusMetrics()
}

// ExecuteSCQuery retrieves data from existing SC trie
func (ef *ElrondNodeFacade) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return ef.apiResolver.ExecuteSCQuery(query)
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (ef *ElrondNodeFacade) PprofEnabled() bool {
	return ef.config.PprofEnabled
}

// IsInterfaceNil returns true if there is no value under the interface
func (ef *ElrondNodeFacade) IsInterfaceNil() bool {
	return ef == nil
}
