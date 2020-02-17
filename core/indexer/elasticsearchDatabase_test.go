package indexer

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/stretchr/testify/require"
)

func newTestElasticSearchDatabase(elasticsearchWriter databaseWriterHandler, arguments elasticSearchDatabaseArgs) *elasticSearchDatabase {
	return &elasticSearchDatabase{
		dbWriter:    elasticsearchWriter,
		marshalizer: arguments.marshalizer,
		hasher:      arguments.hasher,
	}
}

func createMockElasticsearchDatabaseArgs() elasticSearchDatabaseArgs {
	return elasticSearchDatabaseArgs{
		url:         "url",
		userName:    "username",
		password:    "password",
		hasher:      &mock.HasherMock{},
		marshalizer: &mock.MarshalizerMock{},
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

func newTestBlockBody() dataBlock.Body {
	return dataBlock.Body{
		{TxHashes: [][]byte{[]byte("tx1"), []byte("tx2")}, ReceiverShardID: 2, SenderShardID: 2},
		{TxHashes: [][]byte{[]byte("tx3")}, ReceiverShardID: 4, SenderShardID: 1},
	}
}

func TestNewElasticSearchDatabase_IndexesError(t *testing.T) {
	indexes := []string{txIndex, blockIndex, tpsIndex, validatorsIndex, roundIndex}

	for _, index := range indexes {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == ("/" + index) {
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		arguments := createMockElasticsearchDatabaseArgs()
		arguments.url = ts.URL

		elasticDatabase, err := newElasticSearchDatabase(arguments)
		require.Nil(t, elasticDatabase)
		require.Equal(t, ErrCannotCreateIndex, err)
	}
}

func TestElasticseachDatabaseSaveHeader_RequestError(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	localErr := errors.New("localErr")
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			return localErr
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveHeader(header, signerIndexes)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	require.True(t, strings.Contains(output.String(), localErr.Error()))
}

func TestElasticseachDatabaseSaveHeader_CheckRequestBody(t *testing.T) {
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			require.Equal(t, blockIndex, req.Index)

			var block Block
			blockBytes := make([]byte, 227)
			_, _ = req.Body.Read(blockBytes)
			_ = json.Unmarshal(blockBytes, &block)
			require.Equal(t, header.Nonce, block.Nonce)
			require.Equal(t, signerIndexes, block.Validators)

			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveHeader(header, signerIndexes)
}

func TestElasticseachSaveTransactions(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	localErr := errors.New("localErr")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	body := newTestBlockBody()
	header := &dataBlock.Header{Nonce: 1, TxCount: 2}
	txPool := newTestTxPool()

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveTransactions(body, header, txPool, 0)
	require.True(t, strings.Contains(output.String(), "indexing bulk of transactions"))
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	shardId := uint32(0)
	valPubKeys := []string{"key1", "key2"}
	localErr := errors.New("localErr")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			return localErr
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveShardValidatorsPubKeys(shardId, valPubKeys)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	require.True(t, strings.Contains(output.String(), localErr.Error()))
}

func TestElasticsearch_saveShardValidatorsPubKeys(t *testing.T) {
	shardId := uint32(0)
	valPubKeys := []string{"key1", "key2"}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(shardId), 10), req.DocumentID)
			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveShardValidatorsPubKeys(shardId, valPubKeys)
}

func TestElasticsearch_saveShardStatistics_reqError(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	tpsBenchmark := &mock.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	localError := errors.New("local err")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveShardStatistics(tpsBenchmark)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	require.True(t, strings.Contains(output.String(), localError.Error()))
}

func TestElasticsearch_saveShardStatistics(t *testing.T) {
	tpsBenchmark := &mock.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			require.Equal(t, tpsIndex, index)
			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveShardStatistics(tpsBenchmark)
}

func TestElasticsearch_saveRoundInfo(t *testing.T) {
	roundInfo := RoundInfo{
		Index: 1, ShardId: 0, BlockWasProposed: true,
	}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(roundInfo.ShardId), 10)+"_"+strconv.FormatUint(roundInfo.Index, 10), req.DocumentID)
			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveRoundInfo(roundInfo)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	roundInfo := RoundInfo{}
	localError := errors.New("local err")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &mock.DatabaseWriterStub{
		DoRequestCalled: func(req esapi.IndexRequest) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.SaveRoundInfo(roundInfo)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	require.True(t, strings.Contains(output.String(), localError.Error()))
}
