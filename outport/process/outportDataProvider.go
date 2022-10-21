package process

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts/shared"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
)

// ArgOutportDataProvider  holds the arguments needed for creating a new instance of outportDataProvider
type ArgOutportDataProvider struct {
	IsImportDBMode           bool
	ShardCoordinator         sharding.Coordinator
	AlteredAccountsProvider  AlteredAccountsProviderHandler
	TransactionsFeeProcessor TransactionsFeeHandler
	TxCoordinator            process.TransactionCoordinator
	NodesCoordinator         nodesCoordinator.NodesCoordinator
	GasConsumedProvider      GasConsumedProvider
	EconomicsData            EconomicsDataHandler
}

// ArgPrepareOutportSaveBlockData holds the arguments needed for prepare outport save block data
type ArgPrepareOutportSaveBlockData struct {
	HeaderHash             []byte
	Header                 data.HeaderHandler
	Body                   data.BodyHandler
	RewardsTxs             map[string]data.TransactionHandler
	NotarizedHeadersHashes []string
}

type outportDataProvider struct {
	isImportDBMode           bool
	shardID                  uint32
	numOfShards              uint32
	alteredAccountsProvider  AlteredAccountsProviderHandler
	transactionsFeeProcessor TransactionsFeeHandler
	txCoordinator            process.TransactionCoordinator
	nodesCoordinator         nodesCoordinator.NodesCoordinator
	gasConsumedProvider      GasConsumedProvider
	economicsData            EconomicsDataHandler
}

// NewOutportDataProvider will create a new instance of outportDataProvider
func NewOutportDataProvider(arg ArgOutportDataProvider) (*outportDataProvider, error) {
	return &outportDataProvider{
		shardID:                  arg.ShardCoordinator.SelfId(),
		numOfShards:              arg.ShardCoordinator.NumberOfShards(),
		alteredAccountsProvider:  arg.AlteredAccountsProvider,
		transactionsFeeProcessor: arg.TransactionsFeeProcessor,
		txCoordinator:            arg.TxCoordinator,
		nodesCoordinator:         arg.NodesCoordinator,
		gasConsumedProvider:      arg.GasConsumedProvider,
		economicsData:            arg.EconomicsData,
	}, nil
}

// PrepareOutportSaveBlockData will prepare the provided data in a format that will be accepted by an outport driver
func (odp *outportDataProvider) PrepareOutportSaveBlockData(arg ArgPrepareOutportSaveBlockData) (*outportcore.ArgsSaveBlockData, error) {
	if check.IfNil(arg.Header) {
		return nil, errNilHeaderHandler
	}
	if check.IfNil(arg.Body) {
		return nil, errNilBodyHandler
	}

	pool := odp.createPool(arg.RewardsTxs)
	err := odp.transactionsFeeProcessor.PutFeeAndGasUsed(pool)
	if err != nil {
		return nil, fmt.Errorf("transactionsFeeProcessor.PutFeeAndGasUsed %w", err)
	}

	alteredAccounts, err := odp.alteredAccountsProvider.ExtractAlteredAccountsFromPool(pool, shared.AlteredAccountsOptions{})
	if err != nil {
		return nil, fmt.Errorf("alteredAccountsProvider.ExtractAlteredAccountsFromPool %s", err)
	}

	signersIndexes, err := odp.getSignersIndexes(arg.Header)
	if err != nil {
		return nil, err
	}

	return &outportcore.ArgsSaveBlockData{
		HeaderHash:     arg.HeaderHash,
		Body:           arg.Body,
		Header:         arg.Header,
		SignersIndexes: signersIndexes,
		HeaderGasConsumption: outportcore.HeaderGasConsumption{
			GasProvided:    odp.gasConsumedProvider.TotalGasProvidedWithScheduled(),
			GasRefunded:    odp.gasConsumedProvider.TotalGasRefunded(),
			GasPenalized:   odp.gasConsumedProvider.TotalGasPenalized(),
			MaxGasPerBlock: odp.economicsData.MaxGasLimitPerBlock(odp.shardID),
		},
		NotarizedHeadersHashes: arg.NotarizedHeadersHashes,
		TransactionsPool:       pool,
		AlteredAccounts:        alteredAccounts,
		NumberOfShards:         odp.numOfShards,
		IsImportDB:             odp.isImportDBMode,
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

func (odp *outportDataProvider) getSignersIndexes(header data.HeaderHandler) ([]uint64, error) {
	epoch := odp.computeEpoch(header)
	pubKeys, err := odp.nodesCoordinator.GetConsensusValidatorsPublicKeys(
		header.GetPrevRandSeed(),
		header.GetRound(),
		odp.shardID,
		epoch,
	)
	if err != nil {
		return nil, fmt.Errorf("nodesCoordinator.GetConsensusValidatorsPublicKeys %w", err)
	}

	signersIndexes, err := odp.nodesCoordinator.GetValidatorsIndexes(pubKeys, epoch)
	if err != nil {
		return nil, fmt.Errorf("nodesCoordinator.GetValidatorsIndexes %s", err)
	}

	return signersIndexes, nil
}

func (odp *outportDataProvider) createPool(rewardsTxs map[string]data.TransactionHandler) *outportcore.Pool {
	if odp.shardID == core.MetachainShardId {
		return odp.createPoolForMeta(rewardsTxs)
	}

	return odp.createPoolForShard()
}

func (odp *outportDataProvider) createPoolForShard() *outportcore.Pool {
	return &outportcore.Pool{
		Txs:      WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)),
		Scrs:     WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)),
		Rewards:  WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock)),
		Invalid:  WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock)),
		Receipts: WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock)),
		Logs:     odp.txCoordinator.GetAllCurrentLogs(),
	}
}

func (odp *outportDataProvider) createPoolForMeta(rewardsTxs map[string]data.TransactionHandler) *outportcore.Pool {
	return &outportcore.Pool{
		Txs:     WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock)),
		Scrs:    WrapTxsMap(odp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock)),
		Rewards: WrapTxsMap(rewardsTxs),
		Logs:    odp.txCoordinator.GetAllCurrentLogs(),
	}
}

func WrapTxsMap(txs map[string]data.TransactionHandler) map[string]data.TransactionHandlerWithGasUsedAndFee {
	newMap := make(map[string]data.TransactionHandlerWithGasUsedAndFee, len(txs))
	for txHash, tx := range txs {
		newMap[txHash] = outportcore.NewTransactionHandlerWithGasAndFee(tx, 0, big.NewInt(0))
	}

	return newMap
}

// IsInterfaceNil returns true if there is no value under the interface
func (odp *outportDataProvider) IsInterfaceNil() bool {
	return odp == nil
}
