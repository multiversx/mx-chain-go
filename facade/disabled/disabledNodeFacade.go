package disabled

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/api/address"
	_ "github.com/ElrondNetwork/elrond-go/api/block"
	"github.com/ElrondNetwork/elrond-go/api/hardfork"
	"github.com/ElrondNetwork/elrond-go/api/network"
	"github.com/ElrondNetwork/elrond-go/api/node"
	transactionApi "github.com/ElrondNetwork/elrond-go/api/transaction"
	"github.com/ElrondNetwork/elrond-go/api/validator"
	"github.com/ElrondNetwork/elrond-go/api/vmValues"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/statistics"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/data/esdt"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/data/vm"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ = address.FacadeHandler(&disabledNodeFacade{})
var _ = hardfork.FacadeHandler(&disabledNodeFacade{})
var _ = hardfork.FacadeHandler(&disabledNodeFacade{})
var _ = node.FacadeHandler(&disabledNodeFacade{})
var _ = network.FacadeHandler(&disabledNodeFacade{})
var _ = transactionApi.FacadeHandler(&disabledNodeFacade{})
var _ = validator.FacadeHandler(&disabledNodeFacade{})
var _ = vmValues.FacadeHandler(&disabledNodeFacade{})

var errNodeStarting = errors.New("node is starting")
var emptyString = ""

// disabledNodeFacade represents a facade with no functionality
type disabledNodeFacade struct {
	apiInterface         string
	statusMetricsHandler external.StatusMetricsHandler
}

// NewDisabledNodeFacade is the disabled implementation of the facade interface
func NewDisabledNodeFacade(apiInterface string) *disabledNodeFacade {
	return &disabledNodeFacade{
		apiInterface:         apiInterface,
		statusMetricsHandler: NewDisabledStatusMetricsHandler(),
	}
}

// SetSyncer does nothing
func (nf *disabledNodeFacade) SetSyncer(_ ntp.SyncTimer) {
}

// SetTpsBenchmark does nothing
func (nf *disabledNodeFacade) SetTpsBenchmark(_ statistics.TPSBenchmark) {
}

// TpsBenchmark returns nil
func (nf *disabledNodeFacade) TpsBenchmark() statistics.TPSBenchmark {
	return nil
}

// RestAPIServerDebugMode returns false
//TODO: remove in the future
func (nf *disabledNodeFacade) RestAPIServerDebugMode() bool {
	return false
}

// RestApiInterface returns empty string
func (nf *disabledNodeFacade) RestApiInterface() string {
	return nf.apiInterface
}

// GetBalance returns nil and error
func (nf *disabledNodeFacade) GetBalance(_ string) (*big.Int, error) {
	return nil, errNodeStarting
}

// GetUsername returns empty string and error
func (nf *disabledNodeFacade) GetUsername(_ string) (string, error) {
	return emptyString, errNodeStarting
}

// GetValueForKey returns an empty string and error
func (nf *disabledNodeFacade) GetValueForKey(_ string, _ string) (string, error) {
	return emptyString, errNodeStarting
}

// GetESDTBalance returns empty strings and error
func (nf *disabledNodeFacade) GetESDTBalance(_ string, _ string) (string, string, error) {
	return emptyString, emptyString, errNodeStarting
}

// GetAllESDTTokens returns nil and error
func (nf *disabledNodeFacade) GetAllESDTTokens(_ string) (map[string]*esdt.ESDigitalToken, error) {
	return nil, errNodeStarting
}

// CreateTransaction return nil and error
func (nf *disabledNodeFacade) CreateTransaction(
	_ uint64,
	_ string,
	_ string,
	_ []byte,
	_ string,
	_ []byte,
	_ uint64,
	_ uint64,
	_ []byte,
	_ string,
	_ string,
	_ uint32,
	_ uint32) (*transaction.Transaction, []byte, error) {
	return nil, nil, errNodeStarting
}

// ValidateTransaction returns error
func (nf *disabledNodeFacade) ValidateTransaction(_ *transaction.Transaction) error {
	return errNodeStarting
}

// ValidateTransactionForSimulation returns error
func (nf *disabledNodeFacade) ValidateTransactionForSimulation(_ *transaction.Transaction, _ bool) error {
	return errNodeStarting
}

// ValidatorStatisticsApi returns nil and error
func (nf *disabledNodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return nil, errNodeStarting
}

// SendBulkTransactions returns 0 and error
func (nf *disabledNodeFacade) SendBulkTransactions(_ []*transaction.Transaction) (uint64, error) {
	return uint64(0), errNodeStarting
}

// SimulateTransactionExecution returns nil and error
func (nf *disabledNodeFacade) SimulateTransactionExecution(_ *transaction.Transaction) (*transaction.SimulationResults, error) {
	return nil, errNodeStarting
}

// GetTransaction returns nil and error
func (nf *disabledNodeFacade) GetTransaction(_ string, _ bool) (*transaction.ApiTransactionResult, error) {
	return nil, errNodeStarting
}

// ComputeTransactionGasLimit returns 0 and error
func (nf *disabledNodeFacade) ComputeTransactionGasLimit(_ *transaction.Transaction) (*transaction.CostResponse, error) {
	return nil, errNodeStarting
}

// GetAccount returns nil and error
func (nf *disabledNodeFacade) GetAccount(_ string) (state.UserAccountHandler, error) {
	return nil, errNodeStarting
}

// GetCode returns nil and error
func (nf *disabledNodeFacade) GetCode(_ state.UserAccountHandler) []byte {
	return nil
}

// GetHeartbeats returns nil and error
func (nf *disabledNodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return nil, errNodeStarting
}

// StatusMetrics will returns nil
func (nf *disabledNodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return nf.statusMetricsHandler
}

// GetTotalStakedValue returns nil and error
func (nf *disabledNodeFacade) GetTotalStakedValue() (*api.StakeValues, error) {
	return nil, errNodeStarting
}

// ExecuteSCQuery returns nil and error
func (nf *disabledNodeFacade) ExecuteSCQuery(_ *process.SCQuery) (*vm.VMOutputApi, error) {
	return nil, errNodeStarting
}

// PprofEnabled returns false
func (nf *disabledNodeFacade) PprofEnabled() bool {
	return false
}

// Trigger returns error
func (nf *disabledNodeFacade) Trigger(_ uint32, _ bool) error {
	return errNodeStarting
}

// IsSelfTrigger returns false
func (nf *disabledNodeFacade) IsSelfTrigger() bool {
	return false
}

// EncodeAddressPubkey returns empty string and error
func (nf *disabledNodeFacade) EncodeAddressPubkey(_ []byte) (string, error) {
	return emptyString, errNodeStarting
}

// DecodeAddressPubkey returns nil and error
func (nf *disabledNodeFacade) DecodeAddressPubkey(_ string) ([]byte, error) {
	return nil, errNodeStarting
}

// GetQueryHandler returns nil and error
func (nf *disabledNodeFacade) GetQueryHandler(_ string) (debug.QueryHandler, error) {
	return nil, errNodeStarting
}

// GetPeerInfo returns nil and error
func (nf *disabledNodeFacade) GetPeerInfo(_ string) ([]core.QueryP2PPeerInfo, error) {
	return nil, errNodeStarting
}

// GetThrottlerForEndpoint returns nil and false
func (nf *disabledNodeFacade) GetThrottlerForEndpoint(_ string) (core.Throttler, bool) {
	return nil, false
}

// GetBlockByHash return nil and error
func (nf *disabledNodeFacade) GetBlockByHash(_ string, _ bool) (*api.Block, error) {
	return nil, errNodeStarting
}

// GetBlockByNonce returns nil and error
func (nf *disabledNodeFacade) GetBlockByNonce(_ uint64, _ bool) (*api.Block, error) {
	return nil, errNodeStarting
}

// Close returns error
func (nf *disabledNodeFacade) Close() error {
	return errNodeStarting
}

// GetNumCheckpointsFromAccountState returns 0
func (nf *disabledNodeFacade) GetNumCheckpointsFromAccountState() uint32 {
	return uint32(0)
}

// GetNumCheckpointsFromPeerState returns 0
func (nf *disabledNodeFacade) GetNumCheckpointsFromPeerState() uint32 {
	return uint32(0)
}

// GetKeyValuePairs nil map
func (nf *disabledNodeFacade) GetKeyValuePairs(_ string) (map[string]string, error) {
	return nil, nil
}

// GetDirectStakedList returns empty slice
func (nf *disabledNodeFacade) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return make([]*api.DirectStakedValue, 0), nil
}

// GetDelegatorsList returns empty slice
func (nf *disabledNodeFacade) GetDelegatorsList() ([]*api.Delegator, error) {
	return make([]*api.Delegator, 0), nil
}

// GetESDTData returns nil and error
func (nf *disabledNodeFacade) GetESDTData(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
	return nil, errNodeStarting
}

// GetAllIssuedESDTs returns nil and error
func (nf *disabledNodeFacade) GetAllIssuedESDTs(_ string) ([]string, error) {
	return nil, errNodeStarting
}

// IsInterfaceNil returns true if there is no value under the interface
func (nf *disabledNodeFacade) IsInterfaceNil() bool {
	return nf == nil
}
