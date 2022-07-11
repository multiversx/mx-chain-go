package external

import (
	"context"
	"encoding/hex"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/node/external/blockAPI"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ArgNodeApiResolver represents the DTO structure used in the NewNodeApiResolver constructor
type ArgNodeApiResolver struct {
	SCQueryService           SCQueryService
	StatusMetricsHandler     StatusMetricsHandler
	TxCostHandler            TransactionCostHandler
	TotalStakedValueHandler  TotalStakedValueHandler
	DirectStakedListHandler  DirectStakedListHandler
	DelegatedListHandler     DelegatedListHandler
	APITransactionHandler    APITransactionHandler
	APIBlockHandler          blockAPI.APIBlockHandler
	APIInternalBlockHandler  blockAPI.APIInternalBlockHandler
	GenesisNodesSetupHandler sharding.GenesisNodesSetupHandler
	ValidatorPubKeyConverter core.PubkeyConverter
	AccountsParser           genesis.AccountsParser
}

// nodeApiResolver can resolve API requests
type nodeApiResolver struct {
	scQueryService           SCQueryService
	statusMetricsHandler     StatusMetricsHandler
	txCostHandler            TransactionCostHandler
	totalStakedValueHandler  TotalStakedValueHandler
	directStakedListHandler  DirectStakedListHandler
	delegatedListHandler     DelegatedListHandler
	apiTransactionHandler    APITransactionHandler
	apiBlockHandler          blockAPI.APIBlockHandler
	apiInternalBlockHandler  blockAPI.APIInternalBlockHandler
	genesisNodesSetupHandler sharding.GenesisNodesSetupHandler
	validatorPubKeyConverter core.PubkeyConverter
	accountsParser           genesis.AccountsParser
}

// NewNodeApiResolver creates a new nodeApiResolver instance
func NewNodeApiResolver(arg ArgNodeApiResolver) (*nodeApiResolver, error) {
	if check.IfNil(arg.SCQueryService) {
		return nil, ErrNilSCQueryService
	}
	if check.IfNil(arg.StatusMetricsHandler) {
		return nil, ErrNilStatusMetrics
	}
	if check.IfNil(arg.TxCostHandler) {
		return nil, ErrNilTransactionCostHandler
	}
	if check.IfNil(arg.TotalStakedValueHandler) {
		return nil, ErrNilTotalStakedValueHandler
	}
	if check.IfNil(arg.DirectStakedListHandler) {
		return nil, ErrNilDirectStakeListHandler
	}
	if check.IfNil(arg.DelegatedListHandler) {
		return nil, ErrNilDelegatedListHandler
	}
	if check.IfNil(arg.APITransactionHandler) {
		return nil, ErrNilAPITransactionHandler
	}
	if check.IfNil(arg.APIBlockHandler) {
		return nil, ErrNilAPIBlockHandler
	}
	if check.IfNil(arg.APIInternalBlockHandler) {
		return nil, ErrNilAPIInternalBlockHandler
	}
	if check.IfNil(arg.GenesisNodesSetupHandler) {
		return nil, ErrNilGenesisNodesSetupHandler
	}
	if check.IfNil(arg.ValidatorPubKeyConverter) {
		return nil, ErrNilValidatorPubKeyConverter
	}
	if check.IfNil(arg.AccountsParser) {
		return nil, ErrNilAccountsParser
	}

	return &nodeApiResolver{
		scQueryService:           arg.SCQueryService,
		statusMetricsHandler:     arg.StatusMetricsHandler,
		txCostHandler:            arg.TxCostHandler,
		totalStakedValueHandler:  arg.TotalStakedValueHandler,
		directStakedListHandler:  arg.DirectStakedListHandler,
		delegatedListHandler:     arg.DelegatedListHandler,
		apiBlockHandler:          arg.APIBlockHandler,
		apiTransactionHandler:    arg.APITransactionHandler,
		apiInternalBlockHandler:  arg.APIInternalBlockHandler,
		genesisNodesSetupHandler: arg.GenesisNodesSetupHandler,
		validatorPubKeyConverter: arg.ValidatorPubKeyConverter,
		accountsParser:           arg.AccountsParser,
	}, nil
}

// ExecuteSCQuery retrieves data stored in a SC account through a VM
func (nar *nodeApiResolver) ExecuteSCQuery(query *process.SCQuery) (*vmcommon.VMOutput, error) {
	return nar.scQueryService.ExecuteQuery(query)
}

// StatusMetrics returns an implementation of the StatusMetricsHandler interface
func (nar *nodeApiResolver) StatusMetrics() StatusMetricsHandler {
	return nar.statusMetricsHandler
}

// ComputeTransactionGasLimit will calculate how many gas a transaction will consume
func (nar *nodeApiResolver) ComputeTransactionGasLimit(tx *transaction.Transaction) (*transaction.CostResponse, error) {
	return nar.txCostHandler.ComputeTransactionGasLimit(tx)
}

// Close closes all underlying components
func (nar *nodeApiResolver) Close() error {
	return nar.scQueryService.Close()
}

// GetTotalStakedValue will return total staked value
func (nar *nodeApiResolver) GetTotalStakedValue(ctx context.Context) (*api.StakeValues, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue(ctx)
}

// GetDirectStakedList will return the list for the direct staked addresses
func (nar *nodeApiResolver) GetDirectStakedList(ctx context.Context) ([]*api.DirectStakedValue, error) {
	return nar.directStakedListHandler.GetDirectStakedList(ctx)
}

// GetDelegatorsList will return the delegators list
func (nar *nodeApiResolver) GetDelegatorsList(ctx context.Context) ([]*api.Delegator, error) {
	return nar.delegatedListHandler.GetDelegatorsList(ctx)
}

// GetTransaction will return the transaction with the given hash and optionally with results
func (nar *nodeApiResolver) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return nar.apiTransactionHandler.GetTransaction(hash, withResults)
}

// GetTransactionsPool will return a structure containing the transactions pool that is to be returned on API calls
func (nar *nodeApiResolver) GetTransactionsPool(fields string) (*common.TransactionsPoolAPIResponse, error) {
	return nar.apiTransactionHandler.GetTransactionsPool(fields)
}

// GetTransactionsPoolForSender will return a structure containing the transactions for sender that is to be returned on API calls
func (nar *nodeApiResolver) GetTransactionsPoolForSender(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
	return nar.apiTransactionHandler.GetTransactionsPoolForSender(sender, fields)
}

// GetLastPoolNonceForSender will return the last nonce from pool for sender that is to be returned on API calls
func (nar *nodeApiResolver) GetLastPoolNonceForSender(sender string) (uint64, error) {
	return nar.apiTransactionHandler.GetLastPoolNonceForSender(sender)
}

// GetTransactionsPoolNonceGapsForSender will return the nonce gaps from pool for sender, if exists, that is to be returned on API calls
func (nar *nodeApiResolver) GetTransactionsPoolNonceGapsForSender(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
	return nar.apiTransactionHandler.GetTransactionsPoolNonceGapsForSender(sender)
}

// GetBlockByHash will return the block with the given hash and optionally with transactions
func (nar *nodeApiResolver) GetBlockByHash(hash string, options api.BlockQueryOptions) (*api.Block, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return nar.apiBlockHandler.GetBlockByHash(decodedHash, options)
}

// GetBlockByNonce will return the block with the given nonce and optionally with transactions
func (nar *nodeApiResolver) GetBlockByNonce(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
	return nar.apiBlockHandler.GetBlockByNonce(nonce, options)
}

// GetBlockByRound will return the block with the given round and optionally with transactions
func (nar *nodeApiResolver) GetBlockByRound(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
	return nar.apiBlockHandler.GetBlockByRound(round, options)
}

// GetInternalMetaBlockByHash will return a meta block by hash
func (nar *nodeApiResolver) GetInternalMetaBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return nar.apiInternalBlockHandler.GetInternalMetaBlockByHash(format, decodedHash)
}

// GetInternalMetaBlockByNonce will return a meta block by nonce
func (nar *nodeApiResolver) GetInternalMetaBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	return nar.apiInternalBlockHandler.GetInternalMetaBlockByNonce(format, nonce)
}

// GetInternalMetaBlockByRound will return a meta block by round
func (nar *nodeApiResolver) GetInternalMetaBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	return nar.apiInternalBlockHandler.GetInternalMetaBlockByRound(format, round)
}

// GetInternalStartOfEpochMetaBlock will return the start of epoch meta block
// for the specified epoch
func (nar *nodeApiResolver) GetInternalStartOfEpochMetaBlock(format common.ApiOutputFormat, epoch uint32) (interface{}, error) {
	return nar.apiInternalBlockHandler.GetInternalStartOfEpochMetaBlock(format, epoch)
}

// GetInternalShardBlockByHash will return a shard block by hash
func (nar *nodeApiResolver) GetInternalShardBlockByHash(format common.ApiOutputFormat, hash string) (interface{}, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return nar.apiInternalBlockHandler.GetInternalShardBlockByHash(format, decodedHash)
}

// GetInternalShardBlockByNonce will return a shard block by nonce
func (nar *nodeApiResolver) GetInternalShardBlockByNonce(format common.ApiOutputFormat, nonce uint64) (interface{}, error) {
	return nar.apiInternalBlockHandler.GetInternalShardBlockByNonce(format, nonce)
}

// GetInternalShardBlockByRound will return a shard block by round
func (nar *nodeApiResolver) GetInternalShardBlockByRound(format common.ApiOutputFormat, round uint64) (interface{}, error) {
	return nar.apiInternalBlockHandler.GetInternalShardBlockByRound(format, round)
}

// GetInternalMiniBlock will return a shard block by round
func (nar *nodeApiResolver) GetInternalMiniBlock(format common.ApiOutputFormat, hash string, epoch uint32) (interface{}, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return nar.apiInternalBlockHandler.GetInternalMiniBlock(format, decodedHash, epoch)
}

// GetGenesisBalances will return the initial balances of the accounts on genesis time
func (nar *nodeApiResolver) GetGenesisBalances() ([]*common.InitialAccountAPI, error) {
	originalAccounts := nar.accountsParser.InitialAccounts()
	resultedAccounts := make([]*common.InitialAccountAPI, 0, len(originalAccounts))

	for _, acc := range originalAccounts {
		delegationData := common.DelegationDataAPI{}
		if acc.GetDelegationHandler() != nil {
			delegationData.Address = acc.GetDelegationHandler().GetAddress()
			delegationData.Value = bigInToString(acc.GetDelegationHandler().GetValue())
		}
		resultedAccounts = append(resultedAccounts, &common.InitialAccountAPI{
			Address:      acc.GetAddress(),
			Supply:       bigInToString(acc.GetSupply()),
			Balance:      bigInToString(acc.GetBalanceValue()),
			StakingValue: bigInToString(acc.GetStakingValue()),
			Delegation:   delegationData,
		})
	}

	return resultedAccounts, nil
}

func bigInToString(input *big.Int) string {
	if input == nil {
		return "0"
	}

	return input.String()
}

// GetGenesisNodesPubKeys will return genesis nodes public keys by shard
func (nar *nodeApiResolver) GetGenesisNodesPubKeys() (map[uint32][]string, map[uint32][]string) {
	eligibleNodesConfig, waitingNodesConfig := nar.genesisNodesSetupHandler.InitialNodesInfo()
	return nar.getInitialNodesPubKeysBytes(eligibleNodesConfig), nar.getInitialNodesPubKeysBytes(waitingNodesConfig)
}

func (nar *nodeApiResolver) getInitialNodesPubKeysBytes(nodesInfo map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) map[uint32][]string {
	nodesInfoPubkeys := make(map[uint32][]string)

	for shardID, ni := range nodesInfo {
		for i := 0; i < len(ni); i++ {
			nodesInfoPubkeys[shardID] = append(nodesInfoPubkeys[shardID], nar.validatorPubKeyConverter.Encode(ni[i].PubKeyBytes()))
		}
	}

	return nodesInfoPubkeys
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *nodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
