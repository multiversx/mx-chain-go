package initial

import (
	"errors"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/facade"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/ntp"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/txsimulator/data"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
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
func NewInitialNodeFacade(apiInterface string, pprofEnabled bool, statusMetricsHandler external.StatusMetricsHandler) (*initialNodeFacade, error) {
	if check.IfNil(statusMetricsHandler) {
		return nil, facade.ErrNilStatusMetrics
	}

	initialStatusMetrics, err := NewInitialStatusMetricsProvider(statusMetricsHandler)
	if err != nil {
		return nil, err
	}

	return &initialNodeFacade{
		apiInterface:         apiInterface,
		statusMetricsHandler: initialStatusMetrics,
		pprofEnabled:         pprofEnabled,
	}, nil
}

// GetProof -
func (inf *initialNodeFacade) GetProof(_ string, _ string) (*common.GetProofResponse, error) {
	return nil, errNodeStarting
}

// GetProofDataTrie -
func (inf *initialNodeFacade) GetProofDataTrie(_ string, _ string, _ string) (*common.GetProofResponse, *common.GetProofResponse, error) {
	return nil, nil, errNodeStarting
}

// GetProofCurrentRootHash -
func (inf *initialNodeFacade) GetProofCurrentRootHash(_ string) (*common.GetProofResponse, error) {
	return nil, errNodeStarting
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
func (inf *initialNodeFacade) GetBalance(_ string, _ api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetUsername returns empty string and error
func (inf *initialNodeFacade) GetUsername(_ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
	return emptyString, api.BlockInfo{}, errNodeStarting
}

// GetCodeHash returns empty string and error
func (inf *initialNodeFacade) GetCodeHash(_ string, _ api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetValueForKey returns an empty string and error
func (inf *initialNodeFacade) GetValueForKey(_ string, _ string, _ api.AccountQueryOptions) (string, api.BlockInfo, error) {
	return emptyString, api.BlockInfo{}, errNodeStarting
}

// GetESDTBalance returns empty strings and error
func (inf *initialNodeFacade) GetESDTBalance(_ string, _ string, _ api.AccountQueryOptions) (string, string, api.BlockInfo, error) {
	return emptyString, emptyString, api.BlockInfo{}, errNodeStarting
}

// GetAllESDTTokens returns nil and error
func (inf *initialNodeFacade) GetAllESDTTokens(_ string, _ api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetNFTTokenIDsRegisteredByAddress returns nil and error
func (inf *initialNodeFacade) GetNFTTokenIDsRegisteredByAddress(_ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetESDTsWithRole returns nil and error
func (inf *initialNodeFacade) GetESDTsWithRole(_ string, _ string, _ api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// CreateTransaction return nil and error
func (inf *initialNodeFacade) CreateTransaction(_ *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error) {
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
func (inf *initialNodeFacade) ValidatorStatisticsApi() (map[string]*accounts.ValidatorApiResponse, error) {
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
func (inf *initialNodeFacade) GetAccount(_ string, _ api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
	return api.AccountResponse{}, api.BlockInfo{}, errNodeStarting
}

// GetAccounts returns error
func (inf *initialNodeFacade) GetAccounts(_ []string, _ api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetCode returns nil and error
func (inf *initialNodeFacade) GetCode(_ []byte, _ api.AccountQueryOptions) []byte {
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

// StatusMetrics will return nil
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

// GetConnectedPeersRatings returns empty string
func (inf *initialNodeFacade) GetConnectedPeersRatings() string {
	return ""
}

// GetEpochStartDataAPI returns nil and error
func (inf *initialNodeFacade) GetEpochStartDataAPI(_ uint32) (*common.EpochStartDataAPI, error) {
	return nil, errNodeStarting
}

// GetThrottlerForEndpoint returns nil and false
func (inf *initialNodeFacade) GetThrottlerForEndpoint(_ string) (core.Throttler, bool) {
	return nil, false
}

// GetBlockByHash return nil and error
func (inf *initialNodeFacade) GetBlockByHash(_ string, _ api.BlockQueryOptions) (*api.Block, error) {
	return nil, errNodeStarting
}

// GetBlockByNonce returns nil and error
func (inf *initialNodeFacade) GetBlockByNonce(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
	return nil, errNodeStarting
}

// GetBlockByRound returns nil and error
func (inf *initialNodeFacade) GetBlockByRound(_ uint64, _ api.BlockQueryOptions) (*api.Block, error) {
	return nil, errNodeStarting
}

// GetAlteredAccountsForBlock returns nil and error
func (inf *initialNodeFacade) GetAlteredAccountsForBlock(_ api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	return nil, errNodeStarting
}

// GetInternalMetaBlockByHash return nil and error
func (inf *initialNodeFacade) GetInternalMetaBlockByHash(_ common.ApiOutputFormat, _ string) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalMetaBlockByNonce returns nil and error
func (inf *initialNodeFacade) GetInternalMetaBlockByNonce(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalMetaBlockByRound returns nil and error
func (inf *initialNodeFacade) GetInternalMetaBlockByRound(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalStartOfEpochMetaBlock returns nil and error
func (inf *initialNodeFacade) GetInternalStartOfEpochMetaBlock(_ common.ApiOutputFormat, _ uint32) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalStartOfEpochValidatorsInfo returns nil and error
func (inf *initialNodeFacade) GetInternalStartOfEpochValidatorsInfo(_ uint32) ([]*state.ShardValidatorInfo, error) {
	return nil, errNodeStarting
}

// GetInternalShardBlockByHash return nil and error
func (inf *initialNodeFacade) GetInternalShardBlockByHash(_ common.ApiOutputFormat, _ string) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalShardBlockByNonce returns nil and error
func (inf *initialNodeFacade) GetInternalShardBlockByNonce(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalShardBlockByRound returns nil and error
func (inf *initialNodeFacade) GetInternalShardBlockByRound(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
	return nil, errNodeStarting
}

// GetInternalMiniBlockByHash return nil and error
func (inf *initialNodeFacade) GetInternalMiniBlockByHash(_ common.ApiOutputFormat, _ string, _ uint32) (interface{}, error) {
	return nil, errNodeStarting
}

// Close returns error
func (inf *initialNodeFacade) Close() error {
	return errNodeStarting
}

// GetKeyValuePairs nil map
func (inf *initialNodeFacade) GetKeyValuePairs(_ string, _ api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetGuardianData returns error
func (inf *initialNodeFacade) GetGuardianData(_ string, _ api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
	return api.GuardianData{}, api.BlockInfo{}, errNodeStarting
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
func (inf *initialNodeFacade) GetESDTData(_ string, _ string, _ uint64, _ api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetESDTsRoles return nil and error
func (inf *initialNodeFacade) GetESDTsRoles(_ string, _ api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
	return nil, api.BlockInfo{}, errNodeStarting
}

// GetAllIssuedESDTs returns nil and error
func (inf *initialNodeFacade) GetAllIssuedESDTs(_ string) ([]string, error) {
	return nil, errNodeStarting
}

// GetTokenSupply returns nil and error
func (inf *initialNodeFacade) GetTokenSupply(_ string) (*api.ESDTSupply, error) {
	return nil, errNodeStarting
}

// GetGenesisNodesPubKeys returns nil and error
func (inf *initialNodeFacade) GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string, error) {
	return nil, nil, errNodeStarting
}

// GetGenesisBalances returns nil and error
func (inf *initialNodeFacade) GetGenesisBalances() ([]*common.InitialAccountAPI, error) {
	return nil, errNodeStarting
}

// GetTransactionsPool returns a nil structure and error
func (inf *initialNodeFacade) GetTransactionsPool(_ string) (*common.TransactionsPoolAPIResponse, error) {
	return nil, errNodeStarting
}

// GetLastPoolNonceForSender returns nonce 0 and error
func (inf *initialNodeFacade) GetLastPoolNonceForSender(_ string) (uint64, error) {
	return 0, errNodeStarting
}

// GetTransactionsPoolNonceGapsForSender returns a nil structure and error
func (inf *initialNodeFacade) GetTransactionsPoolNonceGapsForSender(_ string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	return nil, errNodeStarting
}

// GetTransactionsPoolForSender returns a nil structure and error
func (inf *initialNodeFacade) GetTransactionsPoolForSender(_, _ string) (*common.TransactionsPoolForSenderApiResponse, error) {
	return nil, errNodeStarting
}

// GetGasConfigs return a nil map and error
func (inf *initialNodeFacade) GetGasConfigs() (map[string]map[string]uint64, error) {
	return nil, errNodeStarting
}

// IsDataTrieMigrated returns false and error
func (inf *initialNodeFacade) IsDataTrieMigrated(_ string, _ api.AccountQueryOptions) (bool, error) {
	return false, errNodeStarting
}

// IsInterfaceNil returns true if there is no value under the interface
func (inf *initialNodeFacade) IsInterfaceNil() bool {
	return inf == nil
}
