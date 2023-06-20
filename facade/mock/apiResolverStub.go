package mock

import (
	"context"

	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
)

// ApiResolverStub -
type ApiResolverStub struct {
	ExecuteSCQueryHandler                       func(query *process.SCQuery) (*vmcommon.VMOutput, error)
	StatusMetricsHandler                        func() external.StatusMetricsHandler
	ComputeTransactionGasLimitHandler           func(tx *transaction.Transaction) (*transaction.CostResponse, error)
	GetTotalStakedValueHandler                  func(ctx context.Context) (*api.StakeValues, error)
	GetDirectStakedListHandler                  func(ctx context.Context) ([]*api.DirectStakedValue, error)
	GetDelegatorsListHandler                    func(ctx context.Context) ([]*api.Delegator, error)
	GetBlockByHashCalled                        func(hash string, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByNonceCalled                       func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetBlockByRoundCalled                       func(round uint64, options api.BlockQueryOptions) (*api.Block, error)
	GetAlteredAccountsForBlockCalled            func(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error)
	GetTransactionHandler                       func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error)
	GetInternalShardBlockByNonceCalled          func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalShardBlockByHashCalled           func(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalShardBlockByRoundCalled          func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMetaBlockByNonceCalled           func(format common.ApiOutputFormat, nonce uint64) (interface{}, error)
	GetInternalMetaBlockByHashCalled            func(format common.ApiOutputFormat, hash string) (interface{}, error)
	GetInternalMetaBlockByRoundCalled           func(format common.ApiOutputFormat, round uint64) (interface{}, error)
	GetInternalMiniBlockCalled                  func(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochMetaBlockCalled      func(format common.ApiOutputFormat, epoch uint32) (interface{}, error)
	GetInternalStartOfEpochValidatorsInfoCalled func(epoch uint32) ([]*state.ShardValidatorInfo, error)
	GetGenesisNodesPubKeysCalled                func() (map[uint32][]string, map[uint32][]string)
	GetTransactionsPoolCalled                   func(fields string) (*common.TransactionsPoolAPIResponse, error)
	GetGenesisBalancesCalled                    func() ([]*common.InitialAccountAPI, error)
	GetTransactionsPoolForSenderCalled          func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error)
	GetLastPoolNonceForSenderCalled             func(sender string) (uint64, error)
	GetTransactionsPoolNonceGapsForSenderCalled func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error)
	GetGasConfigsCalled                         func() map[string]map[string]uint64
	GetManagedKeysCountCalled                   func() int
	GetEligibleManagedKeysCalled                func() ([]string, error)
	GetWaitingManagedKeysCalled                 func() ([]string, error)
}

// GetTransaction -
func (ars *ApiResolverStub) GetTransaction(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
	if ars.GetTransactionHandler != nil {
		return ars.GetTransactionHandler(hash, withEvents)
	}

	return nil, nil
}

// GetBlockByHash -
func (ars *ApiResolverStub) GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error) {
	if ars.GetBlockByHashCalled != nil {
		return ars.GetBlockByHashCalled(hash, options)
	}

	return nil, nil
}

// GetBlockByNonce -
func (ars *ApiResolverStub) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if ars.GetBlockByNonceCalled != nil {
		return ars.GetBlockByNonceCalled(nonce, options)
	}

	return nil, nil
}

// GetBlockByRound -
func (ars *ApiResolverStub) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	if ars.GetBlockByRoundCalled != nil {
		return ars.GetBlockByRoundCalled(round, options)
	}

	return nil, nil
}

// GetAlteredAccountsForBlock -
func (ars *ApiResolverStub) GetAlteredAccountsForBlock(options api.GetAlteredAccountsForBlockOptions) ([]*alteredAccount.AlteredAccount, error) {
	if ars.GetAlteredAccountsForBlockCalled != nil {
		return ars.GetAlteredAccountsForBlockCalled(options)
	}

	return nil, nil
}

// ExecuteSCQuery -
func (ars *ApiResolverStub) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	if ars.ExecuteSCQueryHandler != nil {
		return ars.ExecuteSCQueryHandler(query)
	}

	return nil, nil
}

// StatusMetrics -
func (ars *ApiResolverStub) StatusMetrics() external.StatusMetricsHandler {
	if ars.StatusMetricsHandler != nil {
		return ars.StatusMetricsHandler()
	}

	return nil
}

// ComputeTransactionGasLimit -
func (ars *ApiResolverStub) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	if ars.ComputeTransactionGasLimitHandler != nil {
		return ars.ComputeTransactionGasLimitHandler(tx)
	}

	return nil, nil
}

// GetTotalStakedValue -
func (ars *ApiResolverStub) GetTotalStakedValue(ctx context.Context) (*api.StakeValues, error) {
	if ars.GetTotalStakedValueHandler != nil {
		return ars.GetTotalStakedValueHandler(ctx)
	}

	return nil, nil
}

// GetDirectStakedList -
func (ars *ApiResolverStub) GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error) {
	if ars.GetDirectStakedListHandler != nil {
		return ars.GetDirectStakedListHandler(ctx)
	}

	return nil, nil
}

// GetDelegatorsList -
func (ars *ApiResolverStub) GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error) {
	if ars.GetDelegatorsListHandler != nil {
		return ars.GetDelegatorsListHandler(ctx)
	}

	return nil, nil
}

// GetInternalShardBlockByNonce -
func (ars *ApiResolverStub) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ars.GetInternalShardBlockByNonceCalled != nil {
		return ars.GetInternalShardBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetInternalShardBlockByHash -
func (ars *ApiResolverStub) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if ars.GetInternalShardBlockByHashCalled != nil {
		return ars.GetInternalShardBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalShardBlockByRound -
func (ars *ApiResolverStub) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ars.GetInternalShardBlockByRoundCalled != nil {
		return ars.GetInternalShardBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalMetaBlockByNonce -
func (ars *ApiResolverStub) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	if ars.GetInternalMetaBlockByNonceCalled != nil {
		return ars.GetInternalMetaBlockByNonceCalled(format, nonce)
	}
	return nil, nil
}

// GetTransactionsPool -
func (ars *ApiResolverStub) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	if ars.GetTransactionsPoolCalled != nil {
		return ars.GetTransactionsPoolCalled(fields)
	}

	return nil, nil
}

// GetTransactionsPoolForSender -
func (ars *ApiResolverStub) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	if ars.GetTransactionsPoolForSenderCalled != nil {
		return ars.GetTransactionsPoolForSenderCalled(sender, fields)
	}

	return nil, nil
}

// GetLastPoolNonceForSender -
func (ars *ApiResolverStub) GetLastPoolNonceForSender(sender string) (uint64, error) {
	if ars.GetLastPoolNonceForSenderCalled != nil {
		return ars.GetLastPoolNonceForSenderCalled(sender)
	}

	return 0, nil
}

// GetTransactionsPoolNonceGapsForSender -
func (ars *ApiResolverStub) GetTransactionsPoolNonceGapsForSender(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	if ars.GetTransactionsPoolNonceGapsForSenderCalled != nil {
		return ars.GetTransactionsPoolNonceGapsForSenderCalled(sender, senderAccountNonce)
	}

	return nil, nil
}

// GetInternalMetaBlockByHash -
func (ars *ApiResolverStub) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	if ars.GetInternalMetaBlockByHashCalled != nil {
		return ars.GetInternalMetaBlockByHashCalled(format, hash)
	}
	return nil, nil
}

// GetInternalMetaBlockByRound -
func (ars *ApiResolverStub) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	if ars.GetInternalMetaBlockByRoundCalled != nil {
		return ars.GetInternalMetaBlockByRoundCalled(format, round)
	}
	return nil, nil
}

// GetInternalMiniBlock -
func (ars *ApiResolverStub) GetInternalMiniBlock(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
	if ars.GetInternalMiniBlockCalled != nil {
		return ars.GetInternalMiniBlockCalled(format, hash, epoch)
	}
	return nil, nil
}

// GetInternalStartOfEpochMetaBlock -
func (ars *ApiResolverStub) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	if ars.GetInternalStartOfEpochMetaBlockCalled != nil {
		return ars.GetInternalStartOfEpochMetaBlockCalled(format, epoch)
	}
	return nil, nil
}

// GetGenesisNodesPubKeys -
func (ars *ApiResolverStub) GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string) {
	if ars.GetGenesisNodesPubKeysCalled != nil {
		return ars.GetGenesisNodesPubKeysCalled()
	}
	return nil, nil
}

// GetGenesisBalances -
func (ars *ApiResolverStub) GetGenesisBalances() ([]*common.InitialAccountAPI, error) {
	if ars.GetGenesisBalancesCalled != nil {
		return ars.GetGenesisBalancesCalled()
	}

	return nil, nil
}

// GetGasConfigs -
func (ars *ApiResolverStub) GetGasConfigs() map[string]map[string]uint64 {
	if ars.GetGasConfigsCalled != nil {
		return ars.GetGasConfigsCalled()
	}

	return nil
}

// GetInternalStartOfEpochValidatorsInfo -
func (ars *ApiResolverStub) GetInternalStartOfEpochValidatorsInfo(epoch uint32) ([]*state.ShardValidatorInfo, error) {
	if ars.GetInternalStartOfEpochValidatorsInfoCalled != nil {
		return ars.GetInternalStartOfEpochValidatorsInfoCalled(epoch)
	}

	return nil, nil
}

// GetManagedKeysCount -
func (ars *ApiResolverStub) GetManagedKeysCount() int {
	if ars.GetManagedKeysCountCalled != nil {
		return ars.GetManagedKeysCountCalled()
	}
	return 0
}

// GetEligibleManagedKeys -
func (ars *ApiResolverStub) GetEligibleManagedKeys() ([]string, error) {
	if ars.GetEligibleManagedKeysCalled != nil {
		return ars.GetEligibleManagedKeysCalled()
	}
	return make([]string, 0), nil
}

// GetWaitingManagedKeys -
func (ars *ApiResolverStub) GetWaitingManagedKeys() ([]string, error) {
	if ars.GetWaitingManagedKeysCalled != nil {
		return ars.GetWaitingManagedKeysCalled()
	}
	return make([]string, 0), nil
}

// Close -
func (ars *ApiResolverStub) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ars *ApiResolverStub) IsInterfaceNil() bool {
	return ars == nil
}
