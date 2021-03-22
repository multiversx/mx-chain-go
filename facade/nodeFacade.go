package facade

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/api"
	"github.com/ElrondNetwork/elrond-go/api/address"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/middleware"
	"github.com/ElrondNetwork/elrond-go/api/node"
	transactionApi "github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/core/throttler"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	apiData "github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
)

// DefaultRestInterface is the default interface the rest API will start on if not specified
const DefaultRestInterface = "localhost:8080"

// DefaultRestPortOff is the default value that should be passed if it is desired
//  to start the node without a REST endpoint available
const DefaultRestPortOff = "off"

var _ = address.FacadeHandler(&nodeFacade{})
var _ = hardfork.FacadeHandler(&nodeFacade{})
var _ = node.FacadeHandler(&nodeFacade{})
var _ = transactionApi.FacadeHandler(&nodeFacade{})
var _ = validator.FacadeHandler(&nodeFacade{})
var _ = vmValues.FacadeHandler(&nodeFacade{})

var log = logger.GetOrCreate("facade")

type resetHandler interface {
	Reset()
	IsInterfaceNil() bool
}

// ArgNodeFacade represents the argument for the nodeFacade
type ArgNodeFacade struct {
	Node                   NodeHandler
	ApiResolver            ApiResolver
	TxSimulatorProcessor   TransactionSimulatorProcessor
	RestAPIServerDebugMode bool
	WsAntifloodConfig      config.WebServerAntifloodConfig
	FacadeConfig           config.FacadeConfig
	ApiRoutesConfig        config.ApiRoutesConfig
	AccountsState          state.AccountsAdapter
	PeerState              state.AccountsAdapter
}

// nodeFacade represents a facade for grouping the functionality for the node
type nodeFacade struct {
	node                   NodeHandler
	apiResolver            ApiResolver
	syncer                 ntp.SyncTimer
	tpsBenchmark           *statistics.TpsBenchmark
	txSimulatorProc        TransactionSimulatorProcessor
	config                 config.FacadeConfig
	apiRoutesConfig        config.ApiRoutesConfig
	endpointsThrottlers    map[string]core.Throttler
	wsAntifloodConfig      config.WebServerAntifloodConfig
	restAPIServerDebugMode bool
	accountsState          state.AccountsAdapter
	peerState              state.AccountsAdapter
	ctx                    context.Context
	cancelFunc             func()
}

// NewNodeFacade creates a new Facade with a NodeWrapper
func NewNodeFacade(arg ArgNodeFacade) (*nodeFacade, error) {
	if check.IfNil(arg.Node) {
		return nil, ErrNilNode
	}
	if check.IfNil(arg.ApiResolver) {
		return nil, ErrNilApiResolver
	}
	if check.IfNil(arg.TxSimulatorProcessor) {
		return nil, ErrNilTransactionSimulatorProcessor
	}
	if len(arg.ApiRoutesConfig.APIPackages) == 0 {
		return nil, ErrNoApiRoutesConfig
	}
	if arg.WsAntifloodConfig.SimultaneousRequests == 0 {
		return nil, fmt.Errorf("%w, SimultaneousRequests should not be 0", ErrInvalidValue)
	}
	if arg.WsAntifloodConfig.SameSourceRequests == 0 {
		return nil, fmt.Errorf("%w, SameSourceRequests should not be 0", ErrInvalidValue)
	}
	if arg.WsAntifloodConfig.SameSourceResetIntervalInSec == 0 {
		return nil, fmt.Errorf("%w, SameSourceResetIntervalInSec should not be 0", ErrInvalidValue)
	}
	if check.IfNil(arg.AccountsState) {
		return nil, ErrNilAccountState
	}
	if check.IfNil(arg.PeerState) {
		return nil, ErrNilPeerState
	}

	throttlersMap := computeEndpointsNumGoRoutinesThrottlers(arg.WsAntifloodConfig)

	nf := &nodeFacade{
		node:                   arg.Node,
		apiResolver:            arg.ApiResolver,
		restAPIServerDebugMode: arg.RestAPIServerDebugMode,
		txSimulatorProc:        arg.TxSimulatorProcessor,
		wsAntifloodConfig:      arg.WsAntifloodConfig,
		config:                 arg.FacadeConfig,
		apiRoutesConfig:        arg.ApiRoutesConfig,
		endpointsThrottlers:    throttlersMap,
		accountsState:          arg.AccountsState,
		peerState:              arg.PeerState,
	}
	nf.ctx, nf.cancelFunc = context.WithCancel(context.Background())

	return nf, nil
}

func computeEndpointsNumGoRoutinesThrottlers(webServerAntiFloodConfig config.WebServerAntifloodConfig) map[string]core.Throttler {
	throttlersMap := make(map[string]core.Throttler)
	for _, endpointSetting := range webServerAntiFloodConfig.EndpointsThrottlers {
		newThrottler, err := throttler.NewNumGoRoutinesThrottler(endpointSetting.MaxNumGoRoutines)
		if err != nil {
			log.Warn("error when setting the maximum go routines throttler for endpoint",
				"endpoint", endpointSetting.Endpoint,
				"max go routines", endpointSetting.MaxNumGoRoutines,
				"error", err,
			)
			continue
		}
		throttlersMap[endpointSetting.Endpoint] = newThrottler
	}

	return throttlersMap
}

// SetSyncer sets the current syncer
func (nf *nodeFacade) SetSyncer(syncer ntp.SyncTimer) {
	nf.syncer = syncer
}

// SetTpsBenchmark sets the tps benchmark handler
func (nf *nodeFacade) SetTpsBenchmark(tpsBenchmark *statistics.TpsBenchmark) {
	nf.tpsBenchmark = tpsBenchmark
}

// TpsBenchmark returns the tps benchmark handler
func (nf *nodeFacade) TpsBenchmark() *statistics.TpsBenchmark {
	return nf.tpsBenchmark
}

// StartNode starts the underlying node
func (nf *nodeFacade) StartNode() error {
	return nf.node.StartConsensus()
}

// StartBackgroundServices starts all background services needed for the correct functionality of the node
func (nf *nodeFacade) StartBackgroundServices() {
	go nf.startRest()
}

// RestAPIServerDebugMode return true is debug mode for Rest API is enabled
func (nf *nodeFacade) RestAPIServerDebugMode() bool {
	return nf.restAPIServerDebugMode
}

// RestApiInterface returns the interface on which the rest API should start on, based on the config file provided.
// The API will start on the DefaultRestInterface value unless a correct value is passed or
//  the value is explicitly set to off, in which case it will not start at all
func (nf *nodeFacade) RestApiInterface() string {
	if nf.config.RestApiInterface == "" {
		return DefaultRestInterface
	}

	return nf.config.RestApiInterface
}

func (nf *nodeFacade) startRest() {
	log.Trace("starting REST api server")

	switch nf.RestApiInterface() {
	case DefaultRestPortOff:
		log.Debug("web server is off")
	default:
		log.Debug("creating web server limiters")
		limiters, err := nf.CreateMiddlewareLimiters()
		if err != nil {
			log.Error("error creating web server limiters",
				"error", err.Error(),
			)
			log.Error("web server is off")
			return
		}

		log.Debug("starting web server",
			"SimultaneousRequests", nf.wsAntifloodConfig.SimultaneousRequests,
			"SameSourceRequests", nf.wsAntifloodConfig.SameSourceRequests,
			"SameSourceResetIntervalInSec", nf.wsAntifloodConfig.SameSourceResetIntervalInSec,
		)

		err = api.Start(nf, nf.apiRoutesConfig, limiters...)
		if err != nil {
			log.Error("could not start webserver",
				"error", err.Error(),
			)
		}
	}
}

// CreateMiddlewareLimiters will create the middleware limiters used in web server
func (nf *nodeFacade) CreateMiddlewareLimiters() ([]api.MiddlewareProcessor, error) {
	sourceLimiter, err := middleware.NewSourceThrottler(nf.wsAntifloodConfig.SameSourceRequests)
	if err != nil {
		return nil, err
	}
	go nf.sourceLimiterReset(sourceLimiter)

	globalLimiter, err := middleware.NewGlobalThrottler(nf.wsAntifloodConfig.SimultaneousRequests)
	if err != nil {
		return nil, err
	}

	return []api.MiddlewareProcessor{sourceLimiter, globalLimiter}, nil
}

func (nf *nodeFacade) sourceLimiterReset(reset resetHandler) {
	betweenResetDuration := time.Second * time.Duration(nf.wsAntifloodConfig.SameSourceResetIntervalInSec)
	for {
		select {
		case <-time.After(betweenResetDuration):
			log.Trace("calling reset on WS source limiter")
			reset.Reset()
		case <-nf.ctx.Done():
			log.Debug("closing nodeFacade.sourceLimiterReset go routine")
			return
		}
	}
}

// GetBalance gets the current balance for a specified address
func (nf *nodeFacade) GetBalance(address string) (*big.Int, error) {
	return nf.node.GetBalance(address)
}

// GetUsername gets the username for a specified address
func (nf *nodeFacade) GetUsername(address string) (string, error) {
	return nf.node.GetUsername(address)
}

// GetValueForKey gets the value for a key in a given address
func (nf *nodeFacade) GetValueForKey(address string, key string) (string, error) {
	return nf.node.GetValueForKey(address, key)
}

// GetESDTBalance returns the ESDT balance and if it is frozen
func (nf *nodeFacade) GetESDTBalance(address string, key string) (string, string, error) {
	return nf.node.GetESDTBalance(address, key)
}

// GetKeyValuePairs returns all the key-value pairs under the provided address
func (nf *nodeFacade) GetKeyValuePairs(address string) (map[string]string, error) {
	return nf.node.GetKeyValuePairs(address)
}

// GetAllESDTTokens returns all the esdt tokens for a given address
func (nf *nodeFacade) GetAllESDTTokens(address string) ([]string, error) {
	return nf.node.GetAllESDTTokens(address)
}

// CreateTransaction creates a transaction from all needed fields
func (nf *nodeFacade) CreateTransaction(
	nonce uint64,
	value string,
	receiver string,
	receiverUsername []byte,
	sender string,
	senderUsername []byte,
	gasPrice uint64,
	gasLimit uint64,
	txData []byte,
	signatureHex string,
	chainID string,
	version uint32,
	options uint32,
) (*transaction.Transaction, []byte, error) {

	return nf.node.CreateTransaction(nonce, value, receiver, receiverUsername, sender, senderUsername, gasPrice, gasLimit, txData, signatureHex, chainID, version, options)
}

// ValidateTransaction will validate a transaction
func (nf *nodeFacade) ValidateTransaction(tx *transaction.Transaction) error {
	return nf.node.ValidateTransaction(tx)
}

// ValidateTransactionForSimulation will validate a transaction for the simulation process
func (nf *nodeFacade) ValidateTransactionForSimulation(tx *transaction.Transaction, checkSignature bool) error {
	return nf.node.ValidateTransactionForSimulation(tx, checkSignature)
}

// ValidatorStatisticsApi will return the statistics for all validators
func (nf *nodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return nf.node.ValidatorStatisticsApi()
}

// SendBulkTransactions will send a bulk of transactions on the topic channel
func (nf *nodeFacade) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return nf.node.SendBulkTransactions(txs)
}

// SimulateTransactionExecution will simulate a transaction's execution and will return the results
func (nf *nodeFacade) SimulateTransactionExecution(tx *transaction.Transaction) (*transaction.SimulationResults, error) {
	return nf.txSimulatorProc.ProcessTx(tx)
}

// GetTransaction gets the transaction with a specified hash
func (nf *nodeFacade) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return nf.node.GetTransaction(hash, withResults)
}

// ComputeTransactionGasLimit will estimate how many gas a transaction will consume
func (nf *nodeFacade) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return nf.apiResolver.ComputeTransactionGasLimit(tx)
}

// GetAccount returns an accountResponse containing information
// about the account correlated with provided address
func (nf *nodeFacade) GetAccount(address string) (state.UserAccountHandler, error) {
	return nf.node.GetAccount(address)
}

// GetCode returns the code for the given account
func (nf *nodeFacade) GetCode(account state.UserAccountHandler) []byte {
	return nf.node.GetCode(account)
}

// GetHeartbeats returns the heartbeat status for each public key from initial list or later joined to the network
func (nf *nodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	hbStatus := nf.node.GetHeartbeats()
	if hbStatus == nil {
		return nil, ErrHeartbeatsNotActive
	}

	return hbStatus, nil
}

// StatusMetrics will return the node's status metrics
func (nf *nodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return nf.apiResolver.StatusMetrics()
}

// GetTotalStakedValue will return total staked value
func (nf *nodeFacade) GetTotalStakedValue() (*apiData.StakeValues, error) {
	return nf.apiResolver.GetTotalStakedValue()
}

// ExecuteSCQuery retrieves data from existing SC trie
func (nf *nodeFacade) ExecuteSCQuery(query *process.SCQuery) (*vm.VMOutputApi, error) {
	vmOutput, err := nf.apiResolver.ExecuteSCQuery(query)
	if err != nil {
		return nil, err
	}

	return nf.convertVmOutputToApiResponse(vmOutput), nil
}

// PprofEnabled returns if profiling mode should be active or not on the application
func (nf *nodeFacade) PprofEnabled() bool {
	return nf.config.PprofEnabled
}

// Trigger will trigger a hardfork event
func (nf *nodeFacade) Trigger(epoch uint32, withEarlyEndOfEpoch bool) error {
	return nf.node.DirectTrigger(epoch, withEarlyEndOfEpoch)
}

// IsSelfTrigger returns true if the self public key is the same with the registered public key
func (nf *nodeFacade) IsSelfTrigger() bool {
	return nf.node.IsSelfTrigger()
}

// EncodeAddressPubkey will encode the provided address public key bytes to string
func (nf *nodeFacade) EncodeAddressPubkey(pk []byte) (string, error) {
	return nf.node.EncodeAddressPubkey(pk)
}

// DecodeAddressPubkey will try to decode the provided address public key string
func (nf *nodeFacade) DecodeAddressPubkey(pk string) ([]byte, error) {
	return nf.node.DecodeAddressPubkey(pk)
}

// GetQueryHandler returns the query handler if existing
func (nf *nodeFacade) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return nf.node.GetQueryHandler(name)
}

// GetPeerInfo returns the peer info of a provided pid
func (nf *nodeFacade) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return nf.node.GetPeerInfo(pid)
}

// GetThrottlerForEndpoint returns the throttler for a given endpoint if found
func (nf *nodeFacade) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	throttlerForEndpoint, ok := nf.endpointsThrottlers[endpoint]
	isThrottlerOk := ok && throttlerForEndpoint != nil

	return throttlerForEndpoint, isThrottlerOk
}

// GetBlockByHash return the block for a given hash
func (nf *nodeFacade) GetBlockByHash(hash string, withTxs bool) (*apiData.Block, error) {
	return nf.node.GetBlockByHash(hash, withTxs)
}

// GetBlockByNonce returns the block for a given nonce
func (nf *nodeFacade) GetBlockByNonce(nonce uint64, withTxs bool) (*apiData.Block, error) {
	return nf.node.GetBlockByNonce(nonce, withTxs)
}

// Close will cleanup started go routines
// TODO use this close method
func (nf *nodeFacade) Close() error {
	nf.cancelFunc()

	return nil
}

// GetNumCheckpointsFromAccountState returns the number of checkpoints of the account state
func (nf *nodeFacade) GetNumCheckpointsFromAccountState() uint32 {
	return nf.accountsState.GetNumCheckpoints()
}

// GetNumCheckpointsFromPeerState returns the number of checkpoints of the peer state
func (nf *nodeFacade) GetNumCheckpointsFromPeerState() uint32 {
	return nf.peerState.GetNumCheckpoints()
}

func (nf *nodeFacade) convertVmOutputToApiResponse(input *vmcommon.VMOutput) *vm.VMOutputApi {
	outputAccounts := make(map[string]*vm.OutputAccountApi)
	for key, acc := range input.OutputAccounts {
		outputAddress, err := nf.node.EncodeAddressPubkey(acc.Address)
		if err != nil {
			log.Warn("cannot encode address", "error", err)
			outputAddress = ""
		}

		storageUpdates := make(map[string]*vm.StorageUpdateApi)
		for updateKey, updateVal := range acc.StorageUpdates {
			storageUpdates[hex.EncodeToString([]byte(updateKey))] = &vm.StorageUpdateApi{
				Offset: updateVal.Offset,
				Data:   updateVal.Data,
			}
		}
		outKey := hex.EncodeToString([]byte(key))
		outAcc := &vm.OutputAccountApi{
			Address:        outputAddress,
			Nonce:          acc.Nonce,
			Balance:        acc.Balance,
			BalanceDelta:   acc.BalanceDelta,
			StorageUpdates: storageUpdates,
			Code:           acc.Code,
			CodeMetadata:   acc.CodeMetadata,
		}

		outAcc.OutputTransfers = make([]vm.OutputTransferApi, len(acc.OutputTransfers))
		for i, outTransfer := range acc.OutputTransfers {
			outTransferApi := vm.OutputTransferApi{
				Value:         outTransfer.Value,
				GasLimit:      outTransfer.GasLimit,
				Data:          outTransfer.Data,
				CallType:      outTransfer.CallType,
				SenderAddress: outputAddress,
			}

			if len(outTransfer.SenderAddress) == len(acc.Address) && !bytes.Equal(outTransfer.SenderAddress, acc.Address) {
				senderAddr, errEncode := nf.node.EncodeAddressPubkey(outTransfer.SenderAddress)
				if errEncode != nil {
					log.Warn("cannot encode address", "error", errEncode)
					senderAddr = outputAddress
				}
				outTransferApi.SenderAddress = senderAddr
			}

			outAcc.OutputTransfers[i] = outTransferApi
		}

		outputAccounts[outKey] = outAcc
	}

	logs := make([]*vm.LogEntryApi, 0, len(input.Logs))
	for i := 0; i < len(input.Logs); i++ {
		originalLog := input.Logs[i]
		logAddress, err := nf.node.EncodeAddressPubkey(originalLog.Address)
		if err != nil {
			log.Warn("cannot encode address", "error", err)
			logAddress = ""
		}

		logs[i] = &vm.LogEntryApi{
			Identifier: originalLog.Identifier,
			Address:    logAddress,
			Topics:     originalLog.Topics,
			Data:       originalLog.Data,
		}
	}

	return &vm.VMOutputApi{
		ReturnData:      input.ReturnData,
		ReturnCode:      input.ReturnCode.String(),
		ReturnMessage:   input.ReturnMessage,
		GasRemaining:    input.GasRemaining,
		GasRefund:       input.GasRefund,
		OutputAccounts:  outputAccounts,
		DeletedAccounts: input.DeletedAccounts,
		TouchedAccounts: input.TouchedAccounts,
		Logs:            logs,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (nf *nodeFacade) IsInterfaceNil() bool {
	return nf == nil
}
