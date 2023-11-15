package mock

import (
	"encoding/hex"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/debug"
	"github.com/multiversx/mx-chain-go/heartbeat/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	txSimData "github.com/multiversx/mx-chain-go/process/transactionEvaluator/data"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
)

// FacadeStub is the mock implementation of a node router handler
type FacadeStub struct {
	ShouldErrorStart                            bool
	ShouldErrorStop                             bool
	GetHeartbeatsHandler                        func() ([]data.PubKeyHeartbeat, error)
	GetBalanceCalled                            func(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error)
	GetAccountCalled                            func(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error)
	GetAccountsCalled                           func(addresses []string, options api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error)
	GenerateTransactionHandler                  func(sender string, receiver string, value *big.Int, code string) (*transaction.Transaction, error)
	GetTransactionHandler                       func(hash string, withResults bool) (*transaction.ApiTransactionResult, error)
	CreateTransactionHandler                    func(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error)
	ValidateTransactionHandler                  func(tx *transaction.Transaction) error
	ValidateTransactionForSimulationHandler     func(tx *transaction.Transaction, bypassSignature bool) error
	SendBulkTransactionsHandler                 func(txs []*transaction.Transaction) (uint64, error)
	ExecuteSCQueryHandler                       func(query *process.SCQuery) (*vm.VMOutputApi, api.BlockInfo, error)
	StatusMetricsHandler                        func() external.StatusMetricsHandler
	ValidatorStatisticsHandler                  func() (map[string]*accounts.ValidatorApiResponse, error)
	ComputeTransactionGasLimitHandler           func(tx *transaction.Transaction) (*transaction.CostResponse, error)
	NodeConfigCalled                            func() map[string]interface{}
	GetQueryHandlerCalled                       func(name string) (debug.QueryHandler, error)
	GetValueForKeyCalled                        func(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetGuardianDataCalled                       func(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error)
	GetPeerInfoCalled                           func(pid string) ([]core.QueryP2PPeerInfo, error)
	GetConnectedPeersRatingsOnMainNetworkCalled func() (string, error)
	GetEpochStartDataAPICalled                  func(epoch uint32) (*common.EpochStartDataAPI, error)
	GetThrottlerForEndpointCalled               func(endpoint string) (core.Throttler, bool)
	GetUsernameCalled                           func(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error)
	GetCodeHashCalled                           func(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error)
	GetKeyValuePairsCalled                      func(address string, options api.AccountQueryOptions) (map[string]string, api.BlockInfo, error)
	SimulateTransactionExecutionHandler         func(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error)
	GetESDTDataCalled                           func(address string, key string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error)
	GetAllESDTTokensCalled                      func(address string, options api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error)
	GetESDTsWithRoleCalled                      func(address string, role string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetESDTsRolesCalled                         func(address string, options api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error)
	GetNFTTokenIDsRegisteredByAddressCalled     func(address string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error)
	GetBlockByHashCalled                        func(hash string, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByNonceCalled                       func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlockCalled            func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error)
	GetBlockByRoundCalled                       func(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetInternalShardBlockByNonceCalled          func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHashCalled           func(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRoundCalled          func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonceCalled           func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHashCalled            func(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRoundCalled           func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalStartOfEpochMetaBlockCalled      func(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfoCalled func(epoch uint32) ([]*state.ShardValidatorInfo, error)
	GetInternalMiniBlockByHashCalled            func(format common.ApiOutputFormat, txHash string, epoch uint32) (interface{}, error)
	GetTotalStakedValueHandler                  func() (*api.StakeValues, error)
	GetAllIssuedESDTsCalled                     func(tokenType string) ([]string, error)
	GetDirectStakedListHandler                  func() ([]*api.DirectStakedValue, error)
	GetDelegatorsListHandler                    func() ([]*api.Delegator, error)
	GetProofCalled                              func(string, string) (*common.GetProofResponse, error)
	GetProofCurrentRootHashCalled               func(string) (*common.GetProofResponse, error)
	GetProofDataTrieCalled                      func(string, string, string) (*common.GetProofResponse, *common.GetProofResponse, error)
	VerifyProofCalled                           func(string, string, [][]byte) (bool, error)
	GetTokenSupplyCalled                        func(token string) (*api.ESDTSupply, error)
	GetGenesisNodesPubKeysCalled                func() (map[uint32][]string, map[uint32][]string, error)
	GetGenesisBalancesCalled                    func() ([]*common.InitialAccountAPI, error)
	GetTransactionsPoolCalled                   func(fields string) (*common.TransactionsPoolAPIResponse, error)
	GetTransactionsPoolForSenderCalled          func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSenderCalled             func(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSenderCalled func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	GetGasConfigsCalled                         func() (map[string]map[string]uint64, error)
	RestApiInterfaceCalled                      func() string
	RestAPIServerDebugModeCalled                func() bool
	PprofEnabledCalled                          func() bool
	DecodeAddressPubkeyCalled                   func(pk string) ([]byte, error)
	IsDataTrieMigratedCalled                    func(address string, options api.AccountQueryOptions) (bool, error)
	GetManagedKeysCountCalled                   func() int
	GetManagedKeysCalled                        func() []string
	GetEligibleManagedKeysCalled                func() ([]string, error)
	GetWaitingManagedKeysCalled                 func() ([]string, error)
	GetWaitingEpochsLeftForPublicKeyCalled      func(publicKey string) (uint32, error)
}

// GetTokenSupply -
func (f *FacadeStub) GetTokenSupply(token string) (*api.ESDTSupply, error) {
	if f.GetTokenSupplyCalled != nil {
		return f.GetTokenSupplyCalled(token)
	}

	return nil, nil
}

// GetProof -
func (f *FacadeStub) GetProof(rootHash string, address string) (*common.GetProofResponse, error) {
	if f.GetProofCalled != nil {
		return f.GetProofCalled(rootHash, address)
	}

	return nil, nil
}

// GetProofCurrentRootHash -
func (f *FacadeStub) GetProofCurrentRootHash(address string) (*common.GetProofResponse, error) {
	if f.GetProofCurrentRootHashCalled != nil {
		return f.GetProofCurrentRootHashCalled(address)
	}

	return nil, nil
}

// GetProofDataTrie -
func (f *FacadeStub) GetProofDataTrie(rootHash string, address string, key string) (*common.GetProofResponse, *common.GetProofResponse, error) {
	if f.GetProofDataTrieCalled != nil {
		return f.GetProofDataTrieCalled(rootHash, address, key)
	}

	return nil, nil, nil
}

// VerifyProof -
func (f *FacadeStub) VerifyProof(rootHash string, address string, proof [][]byte) (bool, error) {
	if f.VerifyProofCalled != nil {
		return f.VerifyProofCalled(rootHash, address, proof)
	}

	return false, nil
}

// GetUsername -
func (f *FacadeStub) GetUsername(address string, options api.AccountQueryOptions) (string, api.BlockInfo, error) {
	if f.GetUsernameCalled != nil {
		return f.GetUsernameCalled(address, options)
	}

	return "", api.BlockInfo{}, nil
}

// GetCodeHash -
func (f *FacadeStub) GetCodeHash(address string, options api.AccountQueryOptions) ([]byte, api.BlockInfo, error) {
	if f.GetCodeHashCalled != nil {
		return f.GetCodeHashCalled(address, options)
	}

	return nil, api.BlockInfo{}, nil
}

// GetThrottlerForEndpoint -
func (f *FacadeStub) GetThrottlerForEndpoint(endpoint string) (core.Throttler, bool) {
	if f.GetThrottlerForEndpointCalled != nil {
		return f.GetThrottlerForEndpointCalled(endpoint)
	}

	return nil, false
}

// RestApiInterface -
func (f *FacadeStub) RestApiInterface() string {
	if f.RestApiInterfaceCalled != nil {
		return f.RestApiInterfaceCalled()
	}
	return "localhost:8080"
}

// RestAPIServerDebugMode -
func (f *FacadeStub) RestAPIServerDebugMode() bool {
	if f.RestAPIServerDebugModeCalled != nil {
		return f.RestAPIServerDebugModeCalled()
	}
	return false
}

// PprofEnabled -
func (f *FacadeStub) PprofEnabled() bool {
	if f.PprofEnabledCalled != nil {
		return f.PprofEnabledCalled()
	}
	return false
}

// GetHeartbeats returns the slice of heartbeat info
func (f *FacadeStub) GetHeartbeats() ([]data.PubKeyHeartbeat, error) {
	return f.GetHeartbeatsHandler()
}

// GetBalance is the mock implementation of a handler's GetBalance method
func (f *FacadeStub) GetBalance(address string, options api.AccountQueryOptions) (*big.Int, api.BlockInfo, error) {
	return f.GetBalanceCalled(address, options)
}

// GetValueForKey is the mock implementation of a handler's GetValueForKey method
func (f *FacadeStub) GetValueForKey(address string, key string, options api.AccountQueryOptions) (string, api.BlockInfo, error) {
	if f.GetValueForKeyCalled != nil {
		return f.GetValueForKeyCalled(address, key, options)
	}

	return "", api.BlockInfo{}, nil
}

// GetKeyValuePairs -
func (f *FacadeStub) GetKeyValuePairs(address string, options api.AccountQueryOptions) (map[string]string, api.BlockInfo, error) {
	if f.GetKeyValuePairsCalled != nil {
		return f.GetKeyValuePairsCalled(address, options)
	}

	return nil, api.BlockInfo{}, nil
}

// GetGuardianData -
func (f *FacadeStub) GetGuardianData(address string, options api.AccountQueryOptions) (api.GuardianData, api.BlockInfo, error) {
	if f.GetGuardianDataCalled != nil {
		return f.GetGuardianDataCalled(address, options)
	}
	return api.GuardianData{}, api.BlockInfo{}, nil
}

// GetESDTData -
func (f *FacadeStub) GetESDTData(address string, key string, nonce uint64, options api.AccountQueryOptions) (*esdt.ESDigitalToken, api.BlockInfo, error) {
	if f.GetESDTDataCalled != nil {
		return f.GetESDTDataCalled(address, key, nonce, options)
	}

	return &esdt.ESDigitalToken{Value: big.NewInt(0)}, api.BlockInfo{}, nil
}

// GetESDTsRoles -
func (f *FacadeStub) GetESDTsRoles(address string, options api.AccountQueryOptions) (map[string][]string, api.BlockInfo, error) {
	if f.GetESDTsRolesCalled != nil {
		return f.GetESDTsRolesCalled(address, options)
	}

	return map[string][]string{}, api.BlockInfo{}, nil
}

// GetAllESDTTokens -
func (f *FacadeStub) GetAllESDTTokens(address string, options api.AccountQueryOptions) (map[string]*esdt.ESDigitalToken, api.BlockInfo, error) {
	if f.GetAllESDTTokensCalled != nil {
		return f.GetAllESDTTokensCalled(address, options)
	}

	return make(map[string]*esdt.ESDigitalToken), api.BlockInfo{}, nil
}

// GetNFTTokenIDsRegisteredByAddress -
func (f *FacadeStub) GetNFTTokenIDsRegisteredByAddress(address string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
	if f.GetNFTTokenIDsRegisteredByAddressCalled != nil {
		return f.GetNFTTokenIDsRegisteredByAddressCalled(address, options)
	}

	return make([]string, 0), api.BlockInfo{}, nil
}

// GetESDTsWithRole -
func (f *FacadeStub) GetESDTsWithRole(address string, role string, options api.AccountQueryOptions) ([]string, api.BlockInfo, error) {
	if f.GetESDTsWithRoleCalled != nil {
		return f.GetESDTsWithRoleCalled(address, role, options)
	}

	return make([]string, 0), api.BlockInfo{}, nil
}

// GetAllIssuedESDTs -
func (f *FacadeStub) GetAllIssuedESDTs(tokenType string) ([]string, error) {
	if f.GetAllIssuedESDTsCalled != nil {
		return f.GetAllIssuedESDTsCalled(tokenType)
	}

	return make([]string, 0), nil
}

// GetAccount -
func (f *FacadeStub) GetAccount(address string, options api.AccountQueryOptions) (api.AccountResponse, api.BlockInfo, error) {
	return f.GetAccountCalled(address, options)
}

// GetAccounts -
func (f *FacadeStub) GetAccounts(addresses []string, options api.AccountQueryOptions) (map[string]*api.AccountResponse, api.BlockInfo, error) {
	if f.GetAccountsCalled != nil {
		return f.GetAccountsCalled(addresses, options)
	}

	return nil, api.BlockInfo{}, nil
}

// CreateTransaction is  mock implementation of a handler's CreateTransaction method
func (f *FacadeStub) CreateTransaction(txArgs *external.ArgsCreateTransaction) (*transaction.Transaction, []byte, error) {
	return f.CreateTransactionHandler(txArgs)
}

// GetTransaction is the mock implementation of a handler's GetTransaction method
func (f *FacadeStub) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return f.GetTransactionHandler(hash, withResults)
}

// SimulateTransactionExecution is the mock implementation of a handler's SimulateTransactionExecution method
func (f *FacadeStub) SimulateTransactionExecution(tx *transaction.Transaction) (*txSimData.SimulationResultsWithVMOutput, error) {
	return f.SimulateTransactionExecutionHandler(tx)
}

// SendBulkTransactions is the mock implementation of a handler's SendBulkTransactions method
func (f *FacadeStub) SendBulkTransactions(txs []*transaction.Transaction) (uint64, error) {
	return f.SendBulkTransactionsHandler(txs)
}

// ValidateTransaction -
func (f *FacadeStub) ValidateTransaction(tx *transaction.Transaction) error {
	return f.ValidateTransactionHandler(tx)
}

// ValidateTransactionForSimulation -
func (f *FacadeStub) ValidateTransactionForSimulation(tx *transaction.Transaction, bypassSignature bool) error {
	return f.ValidateTransactionForSimulationHandler(tx, bypassSignature)
}

// ValidatorStatisticsApi is the mock implementation of a handler's ValidatorStatisticsApi method
func (f *FacadeStub) ValidatorStatisticsApi() (map[string]*accounts.ValidatorApiResponse, error) {
	return f.ValidatorStatisticsHandler()
}

// ExecuteSCQuery is a mock implementation.
func (f *FacadeStub) ExecuteSCQuery(query *process.SCQuery) (*vm.VMOutputApi, api.BlockInfo, error) {
	return f.ExecuteSCQueryHandler(query)
}

// StatusMetrics is the mock implementation for the StatusMetrics
func (f *FacadeStub) StatusMetrics() external.StatusMetricsHandler {
	return f.StatusMetricsHandler()
}

// GetTotalStakedValue -
func (f *FacadeStub) GetTotalStakedValue() (*api.StakeValues, error) {
	return f.GetTotalStakedValueHandler()
}

// GetDirectStakedList -
func (f *FacadeStub) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return f.GetDirectStakedListHandler()
}

// GetDelegatorsList -
func (f *FacadeStub) GetDelegatorsList() ([]*api.Delegator, error) {
	return f.GetDelegatorsListHandler()
}

// ComputeTransactionGasLimit -
func (f *FacadeStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return f.ComputeTransactionGasLimitHandler(tx)
}

// NodeConfig -
func (f *FacadeStub) NodeConfig() map[string]interface{} {
	return f.NodeConfigCalled()
}

// EncodeAddressPubkey -
func (f *FacadeStub) EncodeAddressPubkey(pk []byte) (string, error) {
	return hex.EncodeToString(pk), nil
}

// DecodeAddressPubkey -
func (f *FacadeStub) DecodeAddressPubkey(pk string) ([]byte, error) {
	if f.DecodeAddressPubkeyCalled != nil {
		return f.DecodeAddressPubkeyCalled(pk)
	}
	return hex.DecodeString(pk)
}

// GetQueryHandler -
func (f *FacadeStub) GetQueryHandler(name string) (debug.QueryHandler, error) {
	return f.GetQueryHandlerCalled(name)
}

// GetPeerInfo -
func (f *FacadeStub) GetPeerInfo(pid string) ([]core.QueryP2PPeerInfo, error) {
	return f.GetPeerInfoCalled(pid)
}

// GetConnectedPeersRatingsOnMainNetwork -
func (f *FacadeStub) GetConnectedPeersRatingsOnMainNetwork() (string, error) {
	return f.GetConnectedPeersRatingsOnMainNetworkCalled()
}

// GetEpochStartDataAPI -
func (f *FacadeStub) GetEpochStartDataAPI(epoch uint32) (*common.EpochStartDataAPI, error) {
	return f.GetEpochStartDataAPICalled(epoch)
}

// GetBlockByNonce -
func (f *FacadeStub) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	return f.GetBlockByNonceCalled(nonce, options)
}

// GetBlockByHash -
func (f *FacadeStub) GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error) {
	return f.GetBlockByHashCalled(hash, options)
}

// GetBlockByRound -
func (f *FacadeStub) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if f.GetBlockByRoundCalled != nil {
		return f.GetBlockByRoundCalled(round, options)
	}
	return nil, nil
}

// GetAlteredAccountsForBlock -
func (f *FacadeStub) GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	if f.GetAlteredAccountsForBlockCalled != nil {
		return f.GetAlteredAccountsForBlockCalled(options)
	}
	return nil, nil
}

// GetInternalMetaBlockByNonce -
func (f *FacadeStub) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if f.GetInternalMetaBlockByNonceCalled != nil {
		return f.GetInternalMetaBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetInternalMetaBlockByHash -
func (f *FacadeStub) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if f.GetInternalMetaBlockByHashCalled != nil {
		return f.GetInternalMetaBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalMetaBlockByRound -
func (f *FacadeStub) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if f.GetInternalMetaBlockByRoundCalled != nil {
		return f.GetInternalMetaBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalShardBlockByNonce -
func (f *FacadeStub) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if f.GetInternalShardBlockByNonceCalled != nil {
		return f.GetInternalShardBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetInternalShardBlockByHash -
func (f *FacadeStub) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if f.GetInternalShardBlockByHashCalled != nil {
		return f.GetInternalShardBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalShardBlockByRound -
func (f *FacadeStub) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if f.GetInternalShardBlockByRoundCalled != nil {
		return f.GetInternalShardBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalStartOfEpochMetaBlock -
func (f *FacadeStub) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	if f.GetInternalStartOfEpochMetaBlockCalled != nil {
		return f.GetInternalStartOfEpochMetaBlockCalled(format, epoch)
	}
	return nil, nil
}

// GetInternalMiniBlockByHash -
func (f *FacadeStub) GetInternalMiniBlockByHash(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
	if f.GetInternalMiniBlockByHashCalled != nil {
		return f.GetInternalMiniBlockByHashCalled(format, hash, epoch)
	}
	return nil, nil
}

// GetGenesisNodesPubKeys -
func (f *FacadeStub) GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string, error) {
	if f.GetGenesisNodesPubKeysCalled != nil {
		return f.GetGenesisNodesPubKeysCalled()
	}
	return nil, nil, nil
}

// GetGenesisBalances -
func (f *FacadeStub) GetGenesisBalances() ([]*common.InitialAccountAPI, error) {
	if f.GetGenesisBalancesCalled != nil {
		return f.GetGenesisBalancesCalled()
	}

	return nil, nil
}

// GetTransactionsPool -
func (f *FacadeStub) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	if f.GetTransactionsPoolCalled != nil {
		return f.GetTransactionsPoolCalled(fields)
	}

	return nil, nil
}

// GetTransactionsPoolForSender -
func (f *FacadeStub) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	if f.GetTransactionsPoolForSenderCalled != nil {
		return f.GetTransactionsPoolForSenderCalled(sender, fields)
	}

	return nil, nil
}

// GetLastPoolNonceForSender -
func (f *FacadeStub) GetLastPoolNonceForSender(sender string) (uint64, error) {
	if f.GetLastPoolNonceForSenderCalled != nil {
		return f.GetLastPoolNonceForSenderCalled(sender)
	}

	return 0, nil
}

// GetTransactionsPoolNonceGapsForSender -
func (f *FacadeStub) GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	if f.GetTransactionsPoolNonceGapsForSenderCalled != nil {
		return f.GetTransactionsPoolNonceGapsForSenderCalled(sender)
	}

	return nil, nil
}

// GetGasConfigs -
func (f *FacadeStub) GetGasConfigs() (map[string]map[string]uint64, error) {
	if f.GetGasConfigsCalled != nil {
		return f.GetGasConfigsCalled()
	}

	return nil, nil
}

// GetInternalStartOfEpochValidatorsInfo -
func (f *FacadeStub) GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error) {
	if f.GetInternalStartOfEpochValidatorsInfoCalled != nil {
		return f.GetInternalStartOfEpochValidatorsInfoCalled(epoch)
	}

	return nil, nil
}

// IsDataTrieMigrated -
func (f *FacadeStub) IsDataTrieMigrated(address string, options api.AccountQueryOptions) (bool, error) {
	if f.IsDataTrieMigratedCalled != nil {
		return f.IsDataTrieMigratedCalled(address, options)
	}

	return false, nil
}

// Trigger -
func (f *FacadeStub) Trigger(_ uint32, _ bool) error {
	return nil
}

// IsSelfTrigger -
func (f *FacadeStub) IsSelfTrigger() bool {
	return false
}

// GetManagedKeysCount -
func (f *FacadeStub) GetManagedKeysCount() int {
	if f.GetManagedKeysCountCalled != nil {
		return f.GetManagedKeysCountCalled()
	}
	return 0
}

// GetManagedKeys -
func (f *FacadeStub) GetManagedKeys() []string {
	if f.GetManagedKeysCalled != nil {
		return f.GetManagedKeysCalled()
	}
	return make([]string, 0)
}

// GetEligibleManagedKeys -
func (f *FacadeStub) GetEligibleManagedKeys() ([]string, error) {
	if f.GetEligibleManagedKeysCalled != nil {
		return f.GetEligibleManagedKeysCalled()
	}
	return make([]string, 0), nil
}

// GetWaitingManagedKeys -
func (f *FacadeStub) GetWaitingManagedKeys() ([]string, error) {
	if f.GetWaitingManagedKeysCalled != nil {
		return f.GetWaitingManagedKeysCalled()
	}
	return make([]string, 0), nil
}

// GetWaitingEpochsLeftForPublicKey -
func (f *FacadeStub) GetWaitingEpochsLeftForPublicKey(publicKey string) (uint32, error) {
	if f.GetWaitingEpochsLeftForPublicKeyCalled != nil {
		return f.GetWaitingEpochsLeftForPublicKeyCalled(publicKey)
	}
	return 0, nil
}

// Close -
func (f *FacadeStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (f *FacadeStub) IsInterfaceNil() bool {
	return f == nil
}

// WrongFacade is a struct that can be used as a wrong implementation of the node router handler
type WrongFacade struct {
}
