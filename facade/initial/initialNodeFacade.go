package initial

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	txSimData "github.com/ElrondNetwork/elrond-go/process/txsimulator/data"
	"github.com/ElrondNetwork/elrond-go/state"
)

var errNodeStarting = errors.New("node is starting")
var emptyString = ""

// initialNodeFacade represents a facade with no functionality
type initialNodeFacade struct {
	apiInterface         string
	statusMetricsHandler external.StatusMetricsHandler
	pprofEnabled         bool
}

// NewInitialNodeFacade is the initial implementation of the facade interface
func NewInitialNodeFacade(apiInterface string, pprofEnabled bool) *initialNodeFacade {
	return &initialNodeFacade{
		apiInterface:         apiInterface,
		statusMetricsHandler: NewDisabledStatusMetricsHandler(),
		pprofEnabled:         pprofEnabled,
	}
}

// GetProof -
func (inf *initialNodeFacade) GetProof(_ string, _ string) (*shared.GetProofResponse, error) {
	return nil, errNodeStarting
}

// GetProofDataTrie -
func (inf *initialNodeFacade) GetProofDataTrie(_ string, _ string, _ string) (*shared.GetProofResponse, *shared.GetProofResponse, error) {
	return nil, nil, errNodeStarting
}

// GetProofCurrentRootHash -
func (inf *initialNodeFacade) GetProofCurrentRootHash(_ string) (*shared.GetProofResponse, error) {
	return nil, nil, errNodeStarting
}

// VerifyProof -
func (inf *initialNodeFacade) VerifyProof(_ string, _ string, _ [][]byte) (bool, error) {
	return false, errNodeStarting
}

// SetSyncer does nothing
func (inf *initialNodeFacade) SetSyncer(_ ntp.SyncTimer) {
}

// RestAPIServerDebugMode returns false
//TODO: remove in the future
func (inf *initialNodeFacade) RestAPIServerDebugMode() bool {
	return false
}

// RestApiInterface returns empty string
func (inf *initialNodeFacade) RestApiInterface() string {
	return inf.apiInterface
}

// GetBalance returns nil and error
func (inf *initialNodeFacade) GetBalance(_ string) (*big.Int, error) {
	return nil, errNodeStarting
}

// GetUsername returns empty string and error
func (inf *initialNodeFacade) GetUsername(_ string) (string, error) {
	return emptyString, errNodeStarting
}

// GetValueForKey returns an empty string and error
func (inf *initialNodeFacade) GetValueForKey(_ string, _ string) (string, error) {
	return emptyString, errNodeStarting
}

// GetESDTBalance returns empty strings and error
func (inf *initialNodeFacade) GetESDTBalance(_ string, _ string) (string, string, error) {
	return emptyString, emptyString, errNodeStarting
}

// GetAllESDTTokens returns nil and error
func (inf *initialNodeFacade) GetAllESDTTokens(_ string) (map[string]*esdt.ESDigitalToken, error) {
	return nil, errNodeStarting
}

// GetNFTTokenIDsRegisteredByAddress returns nil and error
func (inf *initialNodeFacade) GetNFTTokenIDsRegisteredByAddress(_ string) ([]string, error) {
	return nil, errNodeStarting
}

// GetESDTsWithRole returns nil and error
func (inf *initialNodeFacade) GetESDTsWithRole(_ string, _ string) ([]string, error) {
	return nil, errNodeStarting
}

// CreateTransaction return nil and error
func (inf *initialNodeFacade) CreateTransaction(
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
func (inf *initialNodeFacade) ValidateTransaction(_ *transaction.Transaction) error {
	return errNodeStarting
}

// ValidateTransactionForSimulation returns error
func (inf *initialNodeFacade) ValidateTransactionForSimulation(_ *transaction.Transaction, _ bool) error {
	return errNodeStarting
}

// ValidatorStatisticsApi returns nil and error
func (inf *initialNodeFacade) ValidatorStatisticsApi() (map[string]*state.ValidatorApiResponse, error) {
	return nil, errNodeStarting
}

// SendBulkTransactions returns 0 and error
func (inf *initialNodeFacade) SendBulkTransactions(_ []*transaction.Transaction) (uint64, error) {
	return uint64(0), errNodeStarting
}

// SimulateTransactionExecution returns nil and error
func (inf *initialNodeFacade) SimulateTransactionExecution(_ *transaction.Transaction) (*txSimData.SimulationResults, error) {
	return nil, errNodeStarting
}

// GetTransaction returns nil and error
func (inf *initialNodeFacade) GetTransaction(_ string, _ bool) (*transaction.ApiTransactionResult, error) {
	return nil, errNodeStarting
}

// ComputeTransactionGasLimit returns 0 and error
func (inf *initialNodeFacade) ComputeTransactionGasLimit(_ *transaction.Transaction) (*transaction.CostResponse, error) {
	return nil, errNodeStarting
}

// GetAccount returns nil and error
func (inf *initialNodeFacade) GetAccount(_ string) (api.AccountResponse, error) {
	return api.AccountResponse{}, errNodeStarting
}

// GetCode returns nil and error
func (inf *initialNodeFacade) GetCode(_ []byte) []byte {
	return nil
}

// DirectTrigger returns error
func (inf *initialNodeFacade) DirectTrigger(_ uint32, _ bool) error {
	return errNodeStarting
}

// GetHeartbeats returns nil and error
func (inf *initialNodeFacade) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return nil, errNodeStarting
}

// StatusMetrics will returns nil
func (inf *initialNodeFacade) StatusMetrics() external.StatusMetricsHandler {
	return inf.statusMetricsHandler
}

// GetTotalStakedValue returns nil and error
func (inf *initialNodeFacade) GetTotalStakedValue() (*api.StakeValues, error) {
	return nil, errNodeStarting
}

// ExecuteSCQuery returns nil and error
func (inf *initialNodeFacade) ExecuteSCQuery(_ *process.SCQuery) (*vm.VMOutputApi, error) {
	return nil, errNodeStarting
}

// PprofEnabled returns false
func (inf *initialNodeFacade) PprofEnabled() bool {
	return inf.pprofEnabled
}

// Trigger returns error
func (inf *initialNodeFacade) Trigger(_ uint32, _ bool) error {
	return errNodeStarting
}

// IsSelfTrigger returns false
func (inf *initialNodeFacade) IsSelfTrigger() bool {
	return false
}

// EncodeAddressPubkey returns empty string and error
func (inf *initialNodeFacade) EncodeAddressPubkey(_ []byte) (string, error) {
	return emptyString, errNodeStarting
}

// DecodeAddressPubkey returns nil and error
func (inf *initialNodeFacade) DecodeAddressPubkey(_ string) ([]byte, error) {
	return nil, errNodeStarting
}

// GetQueryHandler returns nil and error
func (inf *initialNodeFacade) GetQueryHandler(_ string) (debug.QueryHandler, error) {
	return nil, errNodeStarting
}

// GetPeerInfo returns nil and error
func (inf *initialNodeFacade) GetPeerInfo(_ string) ([]core.QueryP2PPeerInfo, error) {
	return nil, errNodeStarting
}

// GetThrottlerForEndpoint returns nil and false
func (inf *initialNodeFacade) GetThrottlerForEndpoint(_ string) (core.Throttler, bool) {
	return nil, false
}

// GetBlockByHash return nil and error
func (inf *initialNodeFacade) GetBlockByHash(_ string, _ bool) (*api.Block, error) {
	return nil, errNodeStarting
}

// GetBlockByNonce returns nil and error
func (inf *initialNodeFacade) GetBlockByNonce(_ uint64, _ bool) (*api.Block, error) {
	return nil, errNodeStarting
}

// Close returns error
func (inf *initialNodeFacade) Close() error {
	return errNodeStarting
}

// GetNumCheckpointsFromAccountState returns 0
func (inf *initialNodeFacade) GetNumCheckpointsFromAccountState() uint32 {
	return uint32(0)
}

// GetNumCheckpointsFromPeerState returns 0
func (inf *initialNodeFacade) GetNumCheckpointsFromPeerState() uint32 {
	return uint32(0)
}

// GetKeyValuePairs nil map
func (inf *initialNodeFacade) GetKeyValuePairs(_ string) (map[string]string, error) {
	return nil, errNodeStarting
}

// GetDirectStakedList returns empty slice
func (inf *initialNodeFacade) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return nil, errNodeStarting
}

// GetDelegatorsList returns empty slice
func (inf *initialNodeFacade) GetDelegatorsList() ([]*api.Delegator, error) {
	return nil, errNodeStarting
}

// GetESDTData returns nil and error
func (inf *initialNodeFacade) GetESDTData(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
	return nil, errNodeStarting
}

// GetESDTsRoles return nil and error
func (inf *initialNodeFacade) GetESDTsRoles(_ string) (map[string][]string, error) {
	return nil, errNodeStarting
}

// GetAllIssuedESDTs returns nil and error
func (inf *initialNodeFacade) GetAllIssuedESDTs(_ string) ([]string, error) {
	return nil, errNodeStarting
}

// GetTokenSupply returns nil and error
func (inf *initialNodeFacade) GetTokenSupply(_ string) (string, error) {
	return "", errNodeStarting
}

// IsInterfaceNil returns true if there is no value under the interface
func (inf *initialNodeFacade) IsInterfaceNil() bool {
	return inf == nil
}
