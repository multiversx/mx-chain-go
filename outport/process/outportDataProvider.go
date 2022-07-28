package process

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/pkg/errors"
)

type ArgOutportDataProvider struct {
	ShardCoordinator         sharding.Coordinator
	AlteredAccountsProvider  AlteredAccountsProviderHandler
	TransactionsFeeProcessor TransactionsFeeHandler
	TxCoordinator            process.TransactionCoordinator
	NodesCoordinator         nodesCoordinator.NodesCoordinator
	GasConsumedProvider      GasConsumedProvider
	EconomicsData            EconomicsDataHandler
}

type outportDataProvider struct {
	shardID                  uint32
	alteredAccountsProvider  AlteredAccountsProviderHandler
	transactionsFeeProcessor TransactionsFeeHandler
	txCoordinator            process.TransactionCoordinator
	nodesCoordinator         nodesCoordinator.NodesCoordinator
	gasConsumedProvider      GasConsumedProvider
	economicsData            EconomicsDataHandler
}

func NewOutportDataProvider(arg ArgOutportDataProvider) (*outportDataProvider, error) {
	return &outportDataProvider{
		shardID:                  arg.ShardCoordinator.SelfId(),
		alteredAccountsProvider:  arg.AlteredAccountsProvider,
		transactionsFeeProcessor: arg.TransactionsFeeProcessor,
		txCoordinator:            arg.TxCoordinator,
		nodesCoordinator:         arg.NodesCoordinator,
		gasConsumedProvider:      arg.GasConsumedProvider,
		economicsData:            arg.EconomicsData,
	}, nil
}

func (odp *outportDataProvider) PrepareOutportSaveBlockData(
	headerHash []byte,
	body data.BodyHandler,
	header data.HeaderHandler,
	rewardsTxs map[string]data.TransactionHandler,
	notarizedHeadersHashes []string,
) (*indexer.ArgsSaveBlockData, error) {
	epoch := odp.computeEpoch(header)
	pubKeys, err := odp.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		header.GetPrevRandSeed(),
		header.GetRound(),
		odp.shardID,
		epoch,
	)
	if err != nil {
		return nil, errors.Wrap(err, "nodesCoordinator.GetConsensusValidatorsPublicKeys")
	}

	nodesCoordinatorShardID, err := odp.nodesCoordinator.ShardIdForEpoch(epoch)
	if err != nil {
		return nil, errors.Wrap(err, "nodesCoordinator.ShardIdForEpoch")
	}

	if odp.shardID != nodesCoordinatorShardID {
		return nil, fmt.Errorf("different shard id for epoch %d: shardId != nodesCoordinatorShardID", epoch)
	}

	signersIndexes, err := odp.nodesCoordinator.GetValidatorsIndexes(pubKeys, epoch)
	if err != nil {
		return nil, errors.Wrap(err, "nodesCoordinator.GetValidatorsIndexes")
	}

	gasProvidedInHeader := odp.gasConsumedProvider.TotalGasProvidedWithScheduled()
	gasPenalizedInheader := odp.gasConsumedProvider.TotalGasPenalized()
	gasRefundedInHeader := odp.gasConsumedProvider.TotalGasRefunded()
	maxGasInHeader := odp.economicsData.MaxGasLimitPerBlock(odp.shardID)

	pool := odp.createPool(rewardsTxs)
	alteredAccounts, err := odp.alteredAccountsProvider.ExtractAlteredAccountsFromPool(pool)
	if err != nil {
		return nil, errors.Wrap(err, "alteredAccountsProvider.ExtractAlteredAccountsFromPoo")
	}

	err = odp.transactionsFeeProcessor.PutFeeAndGasUsed(pool)
	if err != nil {
		return nil, errors.Wrap(err, "transactionsFeeProcessor.PutFeeAndGasUsed")
	}

	return &indexer.ArgsSaveBlockData{
		HeaderHash:     headerHash,
		Body:           body,
		Header:         header,
		SignersIndexes: signersIndexes,
		HeaderGasConsumption: indexer.HeaderGasConsumption{
			GasProvided:    gasProvidedInHeader,
			GasRefunded:    gasRefundedInHeader,
			GasPenalized:   gasPenalizedInheader,
			MaxGasPerBlock: maxGasInHeader,
		},
		NotarizedHeadersHashes: notarizedHeadersHashes,
		TransactionsPool:       pool,
		AlteredAccounts:        alteredAccounts,
	}, nil
}

func (odp *outportDataProvider) computeEpoch(header data.HeaderHandler) uint32 {
	epoch := header.GetEpoch()
	shouldDecreaseEpoch := header.IsStartOfEpochBlock() && epoch > 0 && odp.shardID != core.MetachainShardId
	if shouldDecreaseEpoch {
		epoch--
	}

	return epoch
}

func (odp *outportDataProvider) createPool(rewardsTxs map[string]data.TransactionHandler) *indexer.Pool {
	pool := &indexer.Pool{
		Txs:  wrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)),
		Scrs: wrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)),
		Logs: odp.txCoordinator.GetAllCurrentLogs(),
	}

	if odp.shardID == core.MetachainShardId {
		pool.Rewards = wrapTxsMap(rewardsTxs)
	} else {
		pool.Rewards = wrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock))
		pool.Invalid = wrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock))
		pool.Receipts = wrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock))
	}

	return pool
}

func wrapTxsMap(txs map[string]data.TransactionHandler) map[string]data.TransactionHandlerWithGasUsedAndFee {
	newMap := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(txs))
	for txHash, tx := range txs {
		newMap[txHash] = indexer.NewTransactionHandlerWithGasAndFee(tx, 0, big.NewInt(0))
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (odp *outportDataProvider) IsInterfaceNil() bool {
	return odp == nil
}
