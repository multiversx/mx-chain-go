package external

import (
	"encoding/hex"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// ArgNodeApiResolver represents the DTO structure used in the NewNodeApiResolver constructor
type ArgNodeApiResolver struct {
	SCQueryService          SCQueryService
	StatusMetricsHandler    StatusMetricsHandler
	TxCostHandler           TransactionCostHandler
	TotalStakedValueHandler TotalStakedValueHandler
	DirectStakedListHandler DirectStakedListHandler
	DelegatedListHandler    DelegatedListHandler
	APITransactionHandler   APITransactionHandler
	APIBlockHandler         APIBlockHandler
}

// nodeApiResolver can resolve API requests
type nodeApiResolver struct {
	scQueryService          SCQueryService
	statusMetricsHandler    StatusMetricsHandler
	txCostHandler           TransactionCostHandler
	totalStakedValueHandler TotalStakedValueHandler
	directStakedListHandler DirectStakedListHandler
	delegatedListHandler    DelegatedListHandler
	apiTransactionHandler   APITransactionHandler
	apiBlockHandler         APIBlockHandler
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

	return &nodeApiResolver{
		scQueryService:          arg.SCQueryService,
		statusMetricsHandler:    arg.StatusMetricsHandler,
		txCostHandler:           arg.TxCostHandler,
		totalStakedValueHandler: arg.TotalStakedValueHandler,
		directStakedListHandler: arg.DirectStakedListHandler,
		delegatedListHandler:    arg.DelegatedListHandler,
		apiBlockHandler:         arg.APIBlockHandler,
		apiTransactionHandler:   arg.APITransactionHandler,
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
func (nar *nodeApiResolver) GetTotalStakedValue() (*api.StakeValues, error) {
	return nar.totalStakedValueHandler.GetTotalStakedValue()
}

// GetDirectStakedList will return the list for the direct staked addresses
func (nar *nodeApiResolver) GetDirectStakedList() ([]*api.DirectStakedValue, error) {
	return nar.directStakedListHandler.GetDirectStakedList()
}

// GetDelegatorsList will return the delegators list
func (nar *nodeApiResolver) GetDelegatorsList() ([]*api.Delegator, error) {
	return nar.delegatedListHandler.GetDelegatorsList()
}

// GetTransaction will return the transaction with the given hash and optionally with results
func (nar *nodeApiResolver) GetTransaction(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
	return nar.apiTransactionHandler.GetTransaction(hash, withResults)
}

// GetBlockByHash will return the block with the given hash and optionally with transactions
func (nar *nodeApiResolver) GetBlockByHash(hash string, withTxs bool) (*api.Block, error) {
	decodedHash, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	return nar.apiBlockHandler.GetBlockByHash(decodedHash, withTxs)
}

// GetBlockByNonce will return the block with the given nonce and optionally with transactions
func (nar *nodeApiResolver) GetBlockByNonce(nonce uint64, withTxs bool) (*api.Block, error) {
	return nar.apiBlockHandler.GetBlockByNonce(nonce, withTxs)
}

// GetBlockByRound will return the block with the given round and optionally with transactions
func (nar *nodeApiResolver) GetBlockByRound(round uint64, withTxs bool) (*api.Block, error) {
	return nar.apiBlockHandler.GetBlockByRound(round, withTxs)
}

// IsInterfaceNil returns true if there is no value under the interface
func (nar *nodeApiResolver) IsInterfaceNil() bool {
	return nar == nil
}
