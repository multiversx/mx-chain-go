package vm

import (
	"bytes"
	"encoding/json"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
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
}

// CreateTestIndexer -
func CreateTestIndexer(
	t *testing.T,
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
		Marshalizer: testMarshalizer,
		Options: &indexer.Options{
			IndexerCacheSize: 100,
			UseKibana:        false,
		},
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

	return ti
}

func (ti *testIndexer) createElasticProcessor(
	shardCoordinator sharding.Coordinator,
	transactionFeeCalculator process.TransactionFeeCalculator,
) indexer.ElasticProcessor {
	databaseClient := ti.createDatabaseClient()

	templatesPath := path.Join("../../../cmd/node/config/elasticIndexTemplates", "noKibana")

	indexTemplates, indexPolicies, _ := indexer.GetElasticTemplatesAndPolicies(templatesPath, false)

	enabledIndexes := []string{"transactions"}
	enabledIndexesMap := make(map[string]struct{})
	for _, index := range enabledIndexes {
		enabledIndexesMap[index] = struct{}{}
	}

	esIndexerArgs := indexer.ArgElasticProcessor{
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
		Marshalizer:              testMarshalizer,
		Hasher:                   testHasher,
		AddressPubkeyConverter:   pubkeyConv,
		ValidatorPubkeyConverter: pubkeyConv,
		Options: &indexer.Options{
			IndexerCacheSize: 100,
			UseKibana:        false,
		},
		DBClient:                 databaseClient,
		EnabledIndexes:           enabledIndexesMap,
		AccountsDB:               &mock.AccountsStub{},
		Denomination:             18,
		TransactionFeeCalculator: transactionFeeCalculator,
		IsInImportDBMode:         false,
		ShardCoordinator:         shardCoordinator,
	}

	esProcessor, _ := indexer.NewElasticProcessor(esIndexerArgs)

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

	txsPool := map[string]data.TransactionHandler{
		string(txHash): tx,
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

		txsPool[string(intTxHash)] = intTx
	}

	header := &block.Header{
		ShardID: ti.shardCoordinator.SelfId(),
	}

	ti.indexer.SaveBlock(blk, header, txsPool, nil, nil, nil)

	// wait --> dataDispatcher needs to save the block
	time.Sleep(5 * time.Millisecond)
}

func (ti *testIndexer) createDatabaseClient() indexer.DatabaseClientHandler {
	doBulkRequest := func(buff *bytes.Buffer, index string) error {
		ti.mutex.Lock()
		defer ti.mutex.Unlock()

		ti.indexerData[index] = buff
		return nil
	}

	dbwm := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: doBulkRequest,
	}

	return dbwm
}

// GetIndexerPreparedTransaction -
func (ti *testIndexer) GetIndexerPreparedTransaction(t *testing.T) *indexer.Transaction {
	ti.mutex.RLock()
	txData, ok := ti.indexerData["transactions"]
	ti.mutex.RUnlock()

	require.True(t, ok)

	split := bytes.Split(txData.Bytes(), []byte("\n"))
	require.True(t, len(split) > 2)

	newTx := &indexer.Transaction{}
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
