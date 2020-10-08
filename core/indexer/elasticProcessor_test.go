package indexer

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestElasticSearchDatabase(elasticsearchWriter DatabaseClientHandler, arguments ArgElasticProcessor) *elasticProcessor {
	return &elasticProcessor{
		txDatabaseProcessor: newTxDatabaseProcessor(
			arguments.Hasher,
			arguments.Marshalizer,
			arguments.AddressPubkeyConverter,
			arguments.ValidatorPubkeyConverter,
			arguments.FeeConfig,
		),
		elasticClient: elasticsearchWriter,
		parser: &dataParser{
			marshalizer: arguments.Marshalizer,
			hasher:      arguments.Hasher,
		},
		enabledIndexes: arguments.EnabledIndexes,
		accountsDB:     arguments.AccountsDB,
	}
}

func createMockElasticProcessorArgs() ArgElasticProcessor {
	return ArgElasticProcessor{
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(32),
		Hasher:                   &mock.HasherMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		DBClient:                 &mock.DatabaseWriterStub{},
		Options:                  &Options{},
		EnabledIndexes: map[string]struct{}{
			blockIndex: {}, txIndex: {}, miniblocksIndex: {}, tpsIndex: {}, validatorsIndex: {}, roundIndex: {}, accountsIndex: {}, ratingIndex: {}, accountsHistoryIndex: {},
		},
		AccountsDB: &mock.AccountsStub{},
		FeeConfig: &config.FeeSettings{
			MinGasLimit:    "10",
			GasPerDataByte: "1",
		},
	}
}

func newTestTxPool() map[string]data.TransactionHandler {
	txPool := map[string]data.TransactionHandler{
		"tx1": &transaction.Transaction{
			Nonce:     uint64(1),
			Value:     big.NewInt(1),
			RcvAddr:   []byte("receiver_address1"),
			SndAddr:   []byte("sender_address1"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data1"),
			Signature: []byte("signature1"),
		},
		"tx2": &transaction.Transaction{
			Nonce:     uint64(2),
			Value:     big.NewInt(2),
			RcvAddr:   []byte("receiver_address2"),
			SndAddr:   []byte("sender_address2"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data2"),
			Signature: []byte("signature2"),
		},
		"tx3": &transaction.Transaction{
			Nonce:     uint64(3),
			Value:     big.NewInt(3),
			RcvAddr:   []byte("receiver_address3"),
			SndAddr:   []byte("sender_address3"),
			GasPrice:  uint64(10000),
			GasLimit:  uint64(1000),
			Data:      []byte("tx_data3"),
			Signature: []byte("signature3"),
		},
	}

	return txPool
}

func newTestBlockBody() *dataBlock.Body {
	return &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")}, ReceiverShardID: 2, SenderShardID: 2},
			{TxHashes: [][]byte{[]byte("tx3")}, ReceiverShardID: 4, SenderShardID: 1},
		},
	}
}

func TestNewElasticProcessorWithKibana(t *testing.T) {
	t.Parallel()

	args := createMockElasticProcessorArgs()
	args.Options = &Options{
		UseKibana: true,
	}
	args.DBClient = &mock.DatabaseWriterStub{}

	elasticProcessor, err := NewElasticProcessor(args)
	require.NoError(t, err)
	require.NotNil(t, elasticProcessor)
}

func TestElasticProcessor_RemoveHeader(t *testing.T) {
	called := false

	args := createMockElasticProcessorArgs()
	args.DBClient = &mock.DatabaseWriterStub{
		DoBulkRemoveCalled: func(index string, hashes []string) error {
			called = true
			return nil
		},
	}

	elasticProcessor, err := NewElasticProcessor(args)
	require.NoError(t, err)

	err = elasticProcessor.RemoveHeader(&dataBlock.Header{})
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticProcessor_RemoveMiniblocks(t *testing.T) {
	called := false

	mb1 := &dataBlock.MiniBlock{Type: dataBlock.PeerBlock}
	mb2 := &dataBlock.MiniBlock{ReceiverShardID: 0, SenderShardID: 1} // should be removed
	mb3 := &dataBlock.MiniBlock{ReceiverShardID: 1, SenderShardID: 1} // should be removed
	mb4 := &dataBlock.MiniBlock{ReceiverShardID: 1, SenderShardID: 0} // should NOT be removed

	args := createMockElasticProcessorArgs()

	mbHash2, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb2)
	mbHash3, _ := core.CalculateHash(args.Marshalizer, args.Hasher, mb3)

	args.DBClient = &mock.DatabaseWriterStub{
		DoBulkRemoveCalled: func(index string, hashes []string) error {
			called = true
			require.Equal(t, hashes[0], hex.EncodeToString(mbHash2))
			require.Equal(t, hashes[1], hex.EncodeToString(mbHash3))
			return nil
		},
	}

	elasticProcessor, err := NewElasticProcessor(args)
	require.NoError(t, err)

	header := &dataBlock.Header{
		ShardID: 1,
		MiniBlockHeaders: []dataBlock.MiniBlockHeader{
			{Hash: []byte("hash1")}, {Hash: []byte("hash2")}, {Hash: []byte("hash3")}, {Hash: []byte("hash4")},
		},
	}
	body := &dataBlock.Body{
		MiniBlocks: dataBlock.MiniBlockSlice{
			mb1, mb2, mb3, mb4,
		},
	}
	err = elasticProcessor.RemoveMiniblocks(header, body)
	require.Nil(t, err)
	require.True(t, called)
}

func TestElasticseachDatabaseSaveHeader_RequestError(t *testing.T) {
	localErr := errors.New("localErr")
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveHeader(header, signerIndexes, &dataBlock.Body{}, nil, 1)
	require.Equal(t, localErr, err)
}

func TestElasticseachDatabaseSaveHeader_CheckRequestBody(t *testing.T) {
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}

	miniBlock := &dataBlock.MiniBlock{Type: dataBlock.TxBlock}
	blockBody := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			miniBlock,
		},
	}

	arguments := createMockElasticProcessorArgs()

	mbHash, _ := core.CalculateHash(arguments.Marshalizer, arguments.Hasher, miniBlock)
	hexEncodedHash := hex.EncodeToString(mbHash)

	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, blockIndex, req.Index)

			var block Block
			blockBytes, _ := ioutil.ReadAll(req.Body)
			_ = json.Unmarshal(blockBytes, &block)
			require.Equal(t, header.Nonce, block.Nonce)
			require.Equal(t, hexEncodedHash, block.MiniBlocksHashes[0])
			require.Equal(t, signerIndexes, block.Validators)

			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	err := elasticDatabase.SaveHeader(header, signerIndexes, blockBody, nil, 1)
	require.Nil(t, err)
}

func TestElasticseachSaveTransactions(t *testing.T) {
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	body := newTestBlockBody()
	header := &dataBlock.Header{Nonce: 1, TxCount: 2}
	txPool := newTestTxPool()

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	err := elasticDatabase.SaveTransactions(body, header, txPool, 0, map[string]bool{})
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveValidatorsRating(t *testing.T) {
	docID := "0_1"
	localErr := errors.New("localErr")

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, docID, req.DocumentID)

			return localErr
		},
	}

	elasticProcessor, _ := NewElasticProcessor(arguments)

	err := elasticProcessor.SaveValidatorsRating(docID, []workItems.ValidatorRatingInfo{
		{PublicKey: "blablabla", Rating: 100},
	})
	require.Equal(t, localErr, err)
}

func TestElasticProcessor_SaveMiniblocks(t *testing.T) {
	localErr := errors.New("localErr")

	arguments := createMockElasticProcessorArgs()
	arguments.DBClient = &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
		DoMultiGetCalled: func(query map[string]interface{}, index string) (map[string]interface{}, error) {
			return nil, nil
		},
	}

	elasticProcessor, _ := NewElasticProcessor(arguments)

	header := &dataBlock.Header{}
	body := &dataBlock.Body{MiniBlocks: dataBlock.MiniBlockSlice{
		{SenderShardID: 0, ReceiverShardID: 1},
	}}
	mbsInDB, err := elasticProcessor.SaveMiniblocks(header, body)
	require.Equal(t, localErr, err)
	require.Equal(t, 0, len(mbsInDB))
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	localErr := errors.New("localErr")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			return localErr
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(shardID, epoch, valPubKeys)
	require.Equal(t, localErr, err)
}

func TestElasticsearch_saveShardValidatorsPubKeys(t *testing.T) {
	shardID := uint32(0)
	epoch := uint32(0)
	valPubKeys := [][]byte{[]byte("key1"), []byte("key2")}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, fmt.Sprintf("%d_%d", shardID, epoch), req.DocumentID)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardValidatorsPubKeys(shardID, epoch, valPubKeys)
	require.Nil(t, err)
}

func TestElasticsearch_saveShardStatistics_reqError(t *testing.T) {
	tpsBenchmark := &testscommon.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	localError := errors.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardStatistics(tpsBenchmark)
	require.Equal(t, localError, err)
}

func TestElasticsearch_saveShardStatistics(t *testing.T) {
	tpsBenchmark := &testscommon.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			require.Equal(t, tpsIndex, index)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveShardStatistics(tpsBenchmark)
	require.Nil(t, err)
}

func TestElasticsearch_saveRoundInfo(t *testing.T) {
	roundInfo := workItems.RoundInfo{
		Index: 1, ShardId: 0, BlockWasProposed: true,
	}
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req *esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(roundInfo.ShardId), 10)+"_"+strconv.FormatUint(roundInfo.Index, 10), req.DocumentID)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo([]workItems.RoundInfo{roundInfo})
	require.Nil(t, err)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	roundInfo := workItems.RoundInfo{}
	localError := errors.New("local err")
	arguments := createMockElasticProcessorArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)

	err := elasticDatabase.SaveRoundsInfo([]workItems.RoundInfo{roundInfo})
	require.Equal(t, localError, err)

}

func TestUpdateMiniBlock(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:       dbClient,
		Marshalizer:    &mock.MarshalizerMock{},
		Hasher:         &mock.HasherMock{},
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)

	header1 := &dataBlock.Header{
		ShardID: 0,
	}
	body1 := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{SenderShardID: 1, ReceiverShardID: 0, TxHashes: [][]byte{[]byte("hash12")}},
			{SenderShardID: 0, ReceiverShardID: 1, TxHashes: [][]byte{[]byte("hash1")}},
		},
	}

	header2 := &dataBlock.Header{
		ShardID: 1,
	}

	// insert
	_, _ = esDatabase.SaveMiniblocks(header1, body1)
	// update
	_, _ = esDatabase.SaveMiniblocks(header2, body1)
}

func TestSaveRoundsInfo(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:       dbClient,
		Marshalizer:    &mock.MarshalizerMock{},
		Hasher:         &mock.HasherMock{},
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)

	roundInfo1 := workItems.RoundInfo{
		Index: 1, ShardId: 0, BlockWasProposed: true,
	}
	roundInfo2 := workItems.RoundInfo{
		Index: 2, ShardId: 0, BlockWasProposed: true,
	}
	roundInfo3 := workItems.RoundInfo{
		Index: 3, ShardId: 0, BlockWasProposed: true,
	}

	_ = esDatabase.SaveRoundsInfo([]workItems.RoundInfo{roundInfo1, roundInfo2, roundInfo3})
}

func TestUpdateTransaction(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:       dbClient,
		Marshalizer:    &mock.MarshalizerMock{},
		Hasher:         &mock.HasherMock{},
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasPrice: 10,
		GasLimit: 500,
	}
	txHash2 := []byte("txHash2")
	sndAddr := []byte("snd")
	tx2 := &transaction.Transaction{
		GasPrice: 10,
		GasLimit: 500,
		SndAddr:  sndAddr,
	}
	txHash3 := []byte("txHash3")
	tx3 := &transaction.Transaction{}

	recHash1 := []byte("recHash1")
	rec1 := &receipt.Receipt{
		Value:  big.NewInt(100),
		TxHash: txHash1,
	}

	scHash1 := []byte("scHash1")
	scResult1 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash1,
	}

	scHash2 := []byte("scHash2")
	scResult2 := &smartContractResult.SmartContractResult{
		OriginalTxHash: txHash2,
		RcvAddr:        sndAddr,
		GasLimit:       500,
		GasPrice:       1,
		Value:          big.NewInt(150),
	}

	rTx1Hash := []byte("rTxHash1")
	rTx1 := &rewardTx.RewardTx{
		Round: 1113,
	}
	rTx2Hash := []byte("rTxHash2")
	rTx2 := &rewardTx.RewardTx{
		Round: 1114,
	}

	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				TxHashes: [][]byte{txHash1, txHash2},
				Type:     dataBlock.TxBlock,
			},
			{
				TxHashes: [][]byte{txHash3},
				Type:     dataBlock.TxBlock,
			},
			{
				Type:     dataBlock.RewardsBlock,
				TxHashes: [][]byte{rTx1Hash, rTx2Hash},
			},
			{
				TxHashes: [][]byte{recHash1},
				Type:     dataBlock.ReceiptBlock,
			},
			{
				TxHashes: [][]byte{scHash1, scHash2},
				Type:     dataBlock.SmartContractResultBlock,
			},
		},
	}
	header := &dataBlock.Header{}
	txPool := map[string]data.TransactionHandler{
		string(txHash1):  tx1,
		string(txHash2):  tx2,
		string(txHash3):  tx3,
		string(recHash1): rec1,
		string(rTx1Hash): rTx1,
		string(rTx2Hash): rTx2,
	}

	body.MiniBlocks[0].ReceiverShardID = 1
	// insert
	_ = esDatabase.SaveTransactions(body, header, txPool, 0, map[string]bool{})

	header.TimeStamp = 1234
	txPool = map[string]data.TransactionHandler{
		string(txHash1): tx1,
		string(txHash2): tx2,
		string(scHash1): scResult1,
		string(scHash2): scResult2,
	}

	// update
	_ = esDatabase.SaveTransactions(body, header, txPool, 1, map[string]bool{})
}

func TestGetMultiple(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:       dbClient,
		Marshalizer:    &mock.MarshalizerMock{},
		Hasher:         &mock.HasherMock{},
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)

	hashes := []string{
		"57cf251084cd7f79563207c52f938359eebdaf27f91fef1335a076f5dc4873351",
		"9a3beb87930e42b820cbcb5e73b224ebfc707308aa377905eda18d4589e2b093",
	}

	es := esDatabase.(*elasticProcessor)
	response, _ := es.getExistingObjMap(hashes, "transactions")
	fmt.Println(response)
}

func TestIndexTransactionDestinationBeforeSourceShard(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:                 dbClient,
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)

	txHash1 := []byte("txHash1")
	tx1 := &transaction.Transaction{
		GasPrice: 10,
		GasLimit: 500,
	}
	txHash2 := []byte("txHash2")
	sndAddr := []byte("snd")
	tx2 := &transaction.Transaction{
		GasPrice: 10,
		GasLimit: 500,
		SndAddr:  sndAddr,
	}

	header := &dataBlock.Header{}
	txPool := map[string]data.TransactionHandler{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}
	body := &dataBlock.Body{
		MiniBlocks: []*dataBlock.MiniBlock{
			{
				TxHashes: [][]byte{txHash1, txHash2},
				Type:     dataBlock.TxBlock,
			},
		},
	}
	body.MiniBlocks[0].ReceiverShardID = 2
	body.MiniBlocks[0].SenderShardID = 1
	isMBSInDB, _ := esDatabase.SaveMiniblocks(header, body)
	_ = esDatabase.SaveTransactions(body, header, txPool, 2, isMBSInDB)

	txPool = map[string]data.TransactionHandler{
		string(txHash1): tx1,
		string(txHash2): tx2,
	}

	header.ShardID = 1
	isMBSInDB, _ = esDatabase.SaveMiniblocks(header, body)
	_ = esDatabase.SaveTransactions(body, header, txPool, 0, isMBSInDB)
}

func TestDoBulkRequestLimit(t *testing.T) {
	t.Skip("test must run only if you have an elasticsearch server on address http://localhost:9200")

	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/"},
	})

	args := ArgElasticProcessor{
		DBClient:                 dbClient,
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
	}

	esDatabase, _ := NewElasticProcessor(args)
	//Generate transaction and hashes
	numTransactions := 1
	dataSize := 900001
	for i := 0; i < 1000; i++ {
		txs, hashes := generateTransactions(numTransactions, dataSize)

		header := &dataBlock.Header{}
		txsPool := make(map[string]data.TransactionHandler)
		for j := 0; j < numTransactions; j++ {
			txsPool[hashes[j]] = &txs[j]
		}

		miniblock := &dataBlock.MiniBlock{
			TxHashes: make([][]byte, numTransactions),
			Type:     dataBlock.TxBlock,
		}
		for j := 0; j < numTransactions; j++ {
			miniblock.TxHashes[j] = []byte(hashes[j])
		}

		body := &dataBlock.Body{
			MiniBlocks: []*dataBlock.MiniBlock{
				miniblock,
			},
		}
		body.MiniBlocks[0].ReceiverShardID = 2
		body.MiniBlocks[0].SenderShardID = 1

		_ = esDatabase.SaveTransactions(body, header, txsPool, 2, map[string]bool{})
	}
}

func TestElasticProcessor_ComputeBalanceAsFloat(t *testing.T) {
	t.Parallel()

	args := createMockElasticProcessorArgs()
	args.Denomination = 10

	epInt, _ := NewElasticProcessor(args)
	require.NotNil(t, epInt)

	ep := epInt.(*elasticProcessor)

	tests := []struct {
		input  *big.Int
		output float64
	}{
		{
			input:  big.NewInt(200000000000000000),
			output: float64(20000000),
		},
		{
			input:  big.NewInt(57777777777),
			output: 5.7777777777,
		},
		{
			input:  big.NewInt(5777779),
			output: 0.0005777779,
		},
		{
			input:  big.NewInt(7),
			output: 0.0000000007,
		},
		{
			input:  big.NewInt(-7),
			output: 0.0,
		},
	}

	for _, tt := range tests {
		out := ep.computeBalanceAsFloat(tt.input)
		assert.Equal(t, tt.output, out)
	}
}
