package vm

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	processIndexer "github.com/ElrondNetwork/elrond-go/core/indexer/process"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/accounts"
	block2 "github.com/ElrondNetwork/elrond-go/core/indexer/process/block"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/generalInfo"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/miniblocks"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/transactions"
	"github.com/ElrondNetwork/elrond-go/core/indexer/process/validators"
	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/integrationTests/mock"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

type testIndexer struct {
	indexer          indexer.Indexer
	indexerData      map[string]*bytes.Buffer
	marshalizer      marshal.Marshalizer
	hasher           hashing.Hasher
	shardCoordinator sharding.Coordinator
	mutex            sync.RWMutex
	saveDoneChan     chan struct{}
	t                testing.TB
}

const timeoutSave = 10 * time.Second

// CreateTestIndexer -
func CreateTestIndexer(
	t testing.TB,
	coordinator sharding.Coordinator,
	economicsDataHandler process.EconomicsDataHandler,
) *testIndexer {
	ti := &testIndexer{
		indexerData: map[string]*bytes.Buffer{},
		mutex:       sync.RWMutex{},
	}

	dispatcher, err := indexer.NewDataDispatcher(100)
	require.Nil(t, err)

	dispatcher.StartIndexData()

	txFeeCalculator, ok := economicsDataHandler.(process.TransactionFeeCalculator)
	require.True(t, ok)

	elasticProcessor := ti.createElasticProcessor(coordinator, txFeeCalculator)

	arguments := indexer.ArgDataIndexer{
		Marshalizer:        testMarshalizer,
		NodesCoordinator:   &mock.NodesCoordinatorMock{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		ShardCoordinator:   coordinator,
		ElasticProcessor:   elasticProcessor,
		DataDispatcher:     dispatcher,
	}

	testIndexer, err := indexer.NewDataIndexer(arguments)
	require.Nil(t, err)

	ti.indexer = testIndexer
	ti.shardCoordinator = coordinator
	ti.marshalizer = testMarshalizer
	ti.hasher = testHasher
	ti.t = t
	ti.saveDoneChan = make(chan struct{})

	return ti
}

func (ti *testIndexer) createElasticProcessor(
	shardCoordinator sharding.Coordinator,
	transactionFeeCalculator process.TransactionFeeCalculator,
) indexer.ElasticProcessor {
	databaseClient := ti.createDatabaseClient()

	templatesPath := "../../../cmd/node/config/elasticIndexTemplates"

	reader := processIndexer.NewTemplatesAndPoliciesReader(templatesPath, false)
	indexTemplates, indexPolicies, _ := reader.GetElasticTemplatesAndPolicies()

	enabledIndexes := []string{"transactions"}
	enabledIndexesMap := make(map[string]struct{})
	for _, index := range enabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}

	esIndexerArgs := &processIndexer.ArgElasticProcessor{
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
		BlockProc:      block2.NewBlockProcessor(testHasher, testMarshalizer),
		MiniblocksProc: miniblocks.NewMiniblocksProcessor(shardCoordinator.SelfId(), testHasher, testMarshalizer),
		AccountsProc:   accounts.NewAccountsProcessor(18, testMarshalizer, pubkeyConv, &mock.AccountsStub{}),
		TxProc: transactions.NewTransactionsProcessor(
			pubkeyConv,
			transactionFeeCalculator,
			false,
			shardCoordinator,
			false,
			disabled.NewNilTxLogsProcessor(),
			testHasher,
			testMarshalizer,
		),
		DBClient:        databaseClient,
		ValidatorsProc:  validators.NewValidatorsProcessor(pubkeyConv),
		GeneralInfoProc: generalInfo.NewGeneralInfoProcessor(),
		EnabledIndexes:  enabledIndexesMap,
	}

	esProcessor, _ := processIndexer.NewElasticProcessor(esIndexerArgs)

	return esProcessor
}

// SaveTransaction -
func (ti *testIndexer) SaveTransaction(
	tx data.TransactionHandler,
	mbType block.Type,
	intermediateTxs []data.TransactionHandler,
) {
	txHash, _ := core.CalculateHash(ti.marshalizer, ti.hasher, tx)

	sndShardID := ti.shardCoordinator.ComputeId(tx.GetSndAddr())
	rcvShardID := ti.shardCoordinator.ComputeId(tx.GetRcvAddr())

	bigTxMb := &block.MiniBlock{
		Type:            mbType,
		TxHashes:        [][]byte{txHash},
		SenderShardID:   sndShardID,
		ReceiverShardID: rcvShardID,
	}

	blk := &block.Body{
		MiniBlocks: []*block.MiniBlock{
			bigTxMb,
		},
	}

	pool := &types.Pool{
		Invalid: make(map[string]data.TransactionHandler),
		Txs:     make(map[string]data.TransactionHandler),
		Scrs:    make(map[string]data.TransactionHandler),
	}

	switch mbType {
	case block.InvalidBlock:
		pool.Invalid[string(txHash)] = tx
	case block.TxBlock:
		pool.Txs[string(txHash)] = tx
	}

	for _, intTx := range intermediateTxs {
		sndShardID = ti.shardCoordinator.ComputeId(intTx.GetSndAddr())
		rcvShardID = ti.shardCoordinator.ComputeId(intTx.GetRcvAddr())

		intTxHash, _ := core.CalculateHash(ti.marshalizer, ti.hasher, intTx)

		mb := &block.MiniBlock{
			Type:            block.SmartContractResultBlock,
			TxHashes:        [][]byte{intTxHash},
			SenderShardID:   sndShardID,
			ReceiverShardID: rcvShardID,
		}

		blk.MiniBlocks = append(blk.MiniBlocks, mb)

		pool.Scrs[string(intTxHash)] = intTx
	}

	header := &block.Header{
		ShardID: ti.shardCoordinator.SelfId(),
	}

	args := &types.ArgsSaveBlockData{
		TransactionsPool: pool,
		Body:             blk,
		Header:           header,
	}
	ti.indexer.SaveBlock(args)

	select {
	case <-ti.saveDoneChan:
		return
	case <-time.After(timeoutSave):
		require.Fail(ti.t, "save indexer item timeout")
	}
}

func (ti *testIndexer) createDatabaseClient() processIndexer.DatabaseClientHandler {
	doBulkRequest := func(buff *bytes.Buffer, index string) error {
		ti.mutex.Lock()
		defer ti.mutex.Unlock()

		ti.indexerData[index] = buff
		ti.saveDoneChan <- struct{}{}
		return nil
	}

	dbwm := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: doBulkRequest,
	}

	return dbwm
}

// GetIndexerPreparedTransaction -
func (ti *testIndexer) GetIndexerPreparedTransaction(t *testing.T) *types.Transaction {
	ti.mutex.RLock()
	txData, ok := ti.indexerData["transactions"]
	ti.mutex.RUnlock()

	require.True(t, ok)

	split := bytes.Split(txData.Bytes(), []byte("\n"))
	require.True(t, len(split) > 2)

	newTx := &types.Transaction{}
	err := json.Unmarshal(split[1], newTx)
	require.Nil(t, err)

	if newTx.Receiver != "" {
		return newTx
	}

	splitAgain := bytes.Split(split[1], []byte(`"upsert":`))
	require.True(t, len(split) > 1)

	ss := splitAgain[1][:len(splitAgain[1])-1]
	err = json.Unmarshal(ss, &newTx)
	require.Nil(t, err)

	return newTx
}
