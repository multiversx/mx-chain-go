package process

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/receipt"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("outport/process/outportDataProvider")

// ArgOutportDataProvider holds the arguments needed for creating a new instance of outportDataProvider
type ArgOutportDataProvider struct {
	IsImportDBMode           bool
	ShardCoordinator         sharding.Coordinator
	AlteredAccountsProvider  AlteredAccountsProviderHandler
	TransactionsFeeProcessor TransactionsFeeHandler
	TxCoordinator            process.TransactionCoordinator
	NodesCoordinator         nodesCoordinator.NodesCoordinator
	GasConsumedProvider      GasConsumedProvider
	EconomicsData            EconomicsDataHandler
	Marshaller               marshal.Marshalizer
	Hasher                   hashing.Hasher
	ExecutionOrderHandler    common.ExecutionOrderGetter
}

// ArgPrepareOutportSaveBlockData holds the arguments needed for prepare outport save block data
type ArgPrepareOutportSaveBlockData struct {
	HeaderHash             []byte
	Header                 data.HeaderHandler
	HeaderBytes            []byte
	HeaderType             string
	Body                   data.BodyHandler
	PreviousHeader         data.HeaderHandler
	RewardsTxs             map[string]data.TransactionHandler
	NotarizedHeadersHashes []string
	HighestFinalBlockNonce uint64
	HighestFinalBlockHash  []byte
}

type outportDataProvider struct {
	shardID                  uint32
	numOfShards              uint32
	alteredAccountsProvider  AlteredAccountsProviderHandler
	transactionsFeeProcessor TransactionsFeeHandler
	txCoordinator            process.TransactionCoordinator
	nodesCoordinator         nodesCoordinator.NodesCoordinator
	gasConsumedProvider      GasConsumedProvider
	economicsData            EconomicsDataHandler
	executionOrderHandler    common.ExecutionOrderGetter
	marshaller               marshal.Marshalizer
	hasher                   hashing.Hasher
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
		executionOrderHandler:    arg.ExecutionOrderHandler,
		marshaller:               arg.Marshaller,
		hasher:                   arg.Hasher,
	}, nil
}

// PrepareOutportSaveBlockData will prepare the provided data in a format that will be accepted by an outport driver
func (odp *outportDataProvider) PrepareOutportSaveBlockData(arg ArgPrepareOutportSaveBlockData) (*outportcore.OutportBlockWithHeaderAndBody, error) {
	if check.IfNil(arg.Header) {
		return nil, ErrNilHeaderHandler
	}
	if check.IfNil(arg.Body) {
		return nil, ErrNilBodyHandler
	}

	pool, err := odp.createPool(arg.RewardsTxs)
	if err != nil {
		return nil, err
	}

	err = odp.transactionsFeeProcessor.PutFeeAndGasUsed(pool)
	if err != nil {
		return nil, fmt.Errorf("transactionsFeeProcessor.PutFeeAndGasUsed %w", err)
	}

	orderedTxHashes, foundTxHashes := odp.setExecutionOrderInTransactionPool(pool)

	executedTxs, err := collectExecutedTxHashes(arg.Body, arg.Header)
	if err != nil {
		log.Warn("PrepareOutportSaveBlockData - collectExecutedTxHashes", "error", err)
	}

	err = checkTxOrder(orderedTxHashes, executedTxs, foundTxHashes)
	if err != nil {
		log.Warn("PrepareOutportSaveBlockData - checkTxOrder", "error", err.Error())
	}

	alteredAccounts, err := odp.alteredAccountsProvider.ExtractAlteredAccountsFromPool(pool, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	if err != nil {
		return nil, fmt.Errorf("alteredAccountsProvider.ExtractAlteredAccountsFromPool %s", err)
	}

	signersIndexes, err := odp.getSignersIndexes(arg.Header)
	if err != nil {
		return nil, err
	}

	intraMiniBlocks, err := odp.getIntraShardMiniBlocks(arg.Body)
	if err != nil {
		return nil, err
	}

	return &outportcore.OutportBlockWithHeaderAndBody{
		OutportBlock: &outportcore.OutportBlock{
			ShardID:         odp.shardID,
			BlockData:       nil, // this will be filled with specific data for each driver
			TransactionPool: pool,
			HeaderGasConsumption: &outportcore.HeaderGasConsumption{
				GasProvided:    odp.gasConsumedProvider.TotalGasProvidedWithScheduled(),
				GasRefunded:    odp.gasConsumedProvider.TotalGasRefunded(),
				GasPenalized:   odp.gasConsumedProvider.TotalGasPenalized(),
				MaxGasPerBlock: odp.economicsData.MaxGasLimitPerBlock(odp.shardID),
			},
			AlteredAccounts:        alteredAccounts,
			NotarizedHeadersHashes: arg.NotarizedHeadersHashes,
			NumberOfShards:         odp.numOfShards,
			SignersIndexes:         signersIndexes,

			HighestFinalBlockNonce: arg.HighestFinalBlockNonce,
			HighestFinalBlockHash:  arg.HighestFinalBlockHash,
		},
		HeaderDataWithBody: &outportcore.HeaderDataWithBody{
			Body:                 arg.Body,
			Header:               arg.Header,
			HeaderHash:           arg.HeaderHash,
			IntraShardMiniBlocks: intraMiniBlocks,
		},
	}, nil
}

func collectExecutedTxHashes(bodyHandler data.BodyHandler, headerHandler data.HeaderHandler) (map[string]struct{}, error) {
	executedTxHashes := make(map[string]struct{})
	mbHeaders := headerHandler.GetMiniBlockHeaderHandlers()
	body, ok := bodyHandler.(*block.Body)
	if !ok {
		return nil, ErrWrongTypeAssertion
	}

	miniBlocks := body.GetMiniBlocks()
	if len(miniBlocks) != len(mbHeaders) {
		return nil, ErrMiniBlocksHeadersMismatch
	}

	var err error
	for i, mbHeader := range mbHeaders {
		err = extractExecutedTxsFromMb(mbHeader, miniBlocks[i], executedTxHashes)
		if err != nil {
			return nil, err
		}
	}

	return executedTxHashes, nil
}

func extractExecutedTxsFromMb(mbHeader data.MiniBlockHeaderHandler, miniBlock *block.MiniBlock, executedTxHashes map[string]struct{}) error {
	if mbHeader == nil {
		return ErrNilMiniBlockHeaderHandler
	}
	if miniBlock == nil {
		return ErrNilMiniBlock
	}
	if executedTxHashes == nil {
		return ErrNilExecutedTxHashes
	}
	if mbHeader.GetTypeInt32() == int32(block.PeerBlock) {
		return nil
	}
	if mbHeader.GetProcessingType() == int32(block.Processed) {
		return nil
	}

	if int(mbHeader.GetIndexOfLastTxProcessed()) > len(miniBlock.TxHashes) {
		return ErrIndexOutOfBounds
	}
	for index := mbHeader.GetIndexOfFirstTxProcessed(); index <= mbHeader.GetIndexOfLastTxProcessed(); index++ {
		txHash := miniBlock.TxHashes[index]
		executedTxHashes[string(txHash)] = struct{}{}
	}

	return nil
}

func (odp *outportDataProvider) setExecutionOrderInTransactionPool(
	pool *outportcore.TransactionPool,
) ([][]byte, int) {
	orderedTxHashes := odp.executionOrderHandler.GetItems()
	if pool == nil {
		return orderedTxHashes, 0
	}

	txGroups := map[string]data.TxWithExecutionOrderHandler{}
	for txHash, txInfo := range pool.Transactions {
		txGroups[txHash] = txInfo
	}
	for scrHash, scrInfo := range pool.SmartContractResults {
		txGroups[scrHash] = scrInfo
	}
	for invalidTxHash, invalidTxInfo := range pool.InvalidTxs {
		txGroups[invalidTxHash] = invalidTxInfo
	}
	for txHash, rewardInfo := range pool.Rewards {
		txGroups[txHash] = rewardInfo
	}

	foundTxHashes := 0
	for order, txHash := range orderedTxHashes {
		tx, found := txGroups[hex.EncodeToString(txHash)]
		if found {
			tx.SetExecutionOrder(uint32(order))
			foundTxHashes++
		}
	}

	return orderedTxHashes, foundTxHashes
}

func checkTxOrder(orderedTxHashes [][]byte, executedTxHashes map[string]struct{}, foundTxHashes int) error {
	if len(orderedTxHashes) > foundTxHashes {
		return fmt.Errorf("%w for numOrderedTx %d, foundTxsInPool %d",
			ErrOrderedTxNotFound, len(orderedTxHashes), foundTxHashes,
		)
	}

	if len(executedTxHashes) == 0 {
		return nil
	}

	return checkBodyTransactionsHaveOrder(orderedTxHashes, executedTxHashes)
}

func checkBodyTransactionsHaveOrder(orderedTxHashes [][]byte, executedTxHashes map[string]struct{}) error {
	if executedTxHashes == nil {
		return ErrNilExecutedTxHashes
	}

	orderedTxHashesMap := make(map[string]struct{})
	for _, txHash := range orderedTxHashes {
		orderedTxHashesMap[string(txHash)] = struct{}{}
	}

	for executedTxHash := range executedTxHashes {
		if _, ok := orderedTxHashesMap[executedTxHash]; !ok {
			return fmt.Errorf("%w for txHash %s", ErrExecutedTxNotFoundInOrderedTxs, hex.EncodeToString([]byte(executedTxHash)))
		}
	}
	return nil
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

func (odp *outportDataProvider) createPool(rewardsTxs map[string]data.TransactionHandler) (*outportcore.TransactionPool, error) {
	if odp.shardID == core.MetachainShardId {
		return odp.createPoolForMeta(rewardsTxs)
	}

	return odp.createPoolForShard()
}

func (odp *outportDataProvider) createPoolForShard() (*outportcore.TransactionPool, error) {
	txs, err := getTxs(odp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock))
	if err != nil {
		return nil, err
	}

	scrs, err := getScrs(odp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock))
	if err != nil {
		return nil, err
	}

	rewards, err := getRewards(odp.txCoordinator.GetAllCurrentUsedTxs(block.RewardsBlock))
	if err != nil {
		return nil, err
	}

	invalidTxs, err := getTxs(odp.txCoordinator.GetAllCurrentUsedTxs(block.InvalidBlock))
	if err != nil {
		return nil, err
	}

	receipts, err := getReceipts(odp.txCoordinator.GetAllCurrentUsedTxs(block.ReceiptBlock))
	if err != nil {
		return nil, err
	}

	logs, err := getLogs(odp.txCoordinator.GetAllCurrentLogs())
	if err != nil {
		return nil, err
	}

	return &outportcore.TransactionPool{
		Transactions:         txs,
		SmartContractResults: scrs,
		Rewards:              rewards,
		InvalidTxs:           invalidTxs,
		Receipts:             receipts,
		Logs:                 logs,
	}, nil
}

func (odp *outportDataProvider) createPoolForMeta(rewardsTxs map[string]data.TransactionHandler) (*outportcore.TransactionPool, error) {
	txs, err := getTxs(odp.txCoordinator.GetAllCurrentUsedTxs(block.TxBlock))
	if err != nil {
		return nil, err
	}

	scrs, err := getScrs(odp.txCoordinator.GetAllCurrentUsedTxs(block.SmartContractResultBlock))
	if err != nil {
		return nil, err
	}

	rewards, err := getRewards(rewardsTxs)
	if err != nil {
		return nil, err
	}

	logs, err := getLogs(odp.txCoordinator.GetAllCurrentLogs())
	if err != nil {
		return nil, err
	}

	return &outportcore.TransactionPool{
		Transactions:         txs,
		SmartContractResults: scrs,
		Rewards:              rewards,
		Logs:                 logs,
	}, nil
}

func getTxs(txs map[string]data.TransactionHandler) (map[string]*outportcore.TxInfo, error) {
	ret := make(map[string]*outportcore.TxInfo, len(txs))

	for txHash, txHandler := range txs {
		tx, castOk := txHandler.(*transaction.Transaction)
		txHashHex := getHexEncodedHash(txHash)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastTransaction, txHashHex)
		}

		ret[txHashHex] = &outportcore.TxInfo{
			Transaction: tx,
			FeeInfo:     newFeeInfo(),
		}
	}

	return ret, nil
}

func getHexEncodedHash(txHash string) string {
	txHashBytes := []byte(txHash)
	return hex.EncodeToString(txHashBytes)
}

func newFeeInfo() *outportcore.FeeInfo {
	return &outportcore.FeeInfo{
		GasUsed:        0,
		Fee:            big.NewInt(0),
		InitialPaidFee: big.NewInt(0),
	}
}

func getScrs(scrs map[string]data.TransactionHandler) (map[string]*outportcore.SCRInfo, error) {
	ret := make(map[string]*outportcore.SCRInfo, len(scrs))

	for scrHash, txHandler := range scrs {
		scr, castOk := txHandler.(*smartContractResult.SmartContractResult)
		scrHashHex := getHexEncodedHash(scrHash)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastSCR, scrHashHex)
		}

		ret[scrHashHex] = &outportcore.SCRInfo{
			SmartContractResult: scr,
			FeeInfo:             newFeeInfo(),
		}
	}

	return ret, nil
}

func getRewards(rewards map[string]data.TransactionHandler) (map[string]*outportcore.RewardInfo, error) {
	ret := make(map[string]*outportcore.RewardInfo, len(rewards))

	for hash, txHandler := range rewards {
		reward, castOk := txHandler.(*rewardTx.RewardTx)
		hexHex := getHexEncodedHash(hash)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastReward, hexHex)
		}

		ret[hexHex] = &outportcore.RewardInfo{
			Reward: reward,
		}
	}

	return ret, nil
}

func getReceipts(receipts map[string]data.TransactionHandler) (map[string]*receipt.Receipt, error) {
	ret := make(map[string]*receipt.Receipt, len(receipts))

	for hash, receiptHandler := range receipts {
		tx, castOk := receiptHandler.(*receipt.Receipt)
		hashHex := getHexEncodedHash(hash)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastReceipt, hashHex)
		}

		ret[hashHex] = tx
	}

	return ret, nil
}

func getLogs(logs []*data.LogData) ([]*outportcore.LogData, error) {
	ret := make([]*outportcore.LogData, len(logs))

	for idx, logData := range logs {
		txHashHex := getHexEncodedHash(logData.TxHash)
		log, castOk := logData.LogHandler.(*transaction.Log)
		if !castOk {
			return nil, fmt.Errorf("%w, hash: %s", errCannotCastLog, txHashHex)
		}

		ret[idx] = &outportcore.LogData{
			TxHash: txHashHex,
			Log:    log,
		}
	}
	return ret, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (odp *outportDataProvider) IsInterfaceNil() bool {
	return odp == nil
}

func (odp *outportDataProvider) getIntraShardMiniBlocks(bodyHandler data.BodyHandler) ([]*block.MiniBlock, error) {
	body, err := outportcore.GetBody(bodyHandler)
	if err != nil {
		return nil, err
	}

	return odp.filterOutDuplicatedMiniBlocks(body.MiniBlocks, odp.txCoordinator.GetCreatedInShardMiniBlocks())
}

func (odp *outportDataProvider) filterOutDuplicatedMiniBlocks(miniBlocksFromBody []*block.MiniBlock, intraMiniBlocks []*block.MiniBlock) ([]*block.MiniBlock, error) {
	filteredMiniBlocks := make([]*block.MiniBlock, 0, len(intraMiniBlocks))
	mapMiniBlocksFromBody := make(map[string]struct{})
	for _, mb := range miniBlocksFromBody {
		mbHash, err := core.CalculateHash(odp.marshaller, odp.hasher, mb)
		if err != nil {
			return nil, err
		}
		mapMiniBlocksFromBody[string(mbHash)] = struct{}{}
	}

	for _, mb := range intraMiniBlocks {
		mbHash, err := core.CalculateHash(odp.marshaller, odp.hasher, mb)
		if err != nil {
			return nil, err
		}

		_, found := mapMiniBlocksFromBody[string(mbHash)]
		if found {
			continue
		}
		filteredMiniBlocks = append(filteredMiniBlocks, mb)
	}

	return filteredMiniBlocks, nil
}
