package indexer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

type databaseWriterMock struct {
	doRequestCalled     func(req esapi.IndexRequest) error
	doBulkRequestCalled func(buff *bytes.Buffer, index string) error
}

func (dwm *databaseWriterMock) doRequest(req esapi.IndexRequest) error {
	if dwm.doRequestCalled != nil {
		return dwm.doRequestCalled(req)
	}
	return nil
}

func (dwm *databaseWriterMock) doBulkRequest(buff *bytes.Buffer, index string) error {
	if dwm.doBulkRequestCalled != nil {
		return dwm.doBulkRequestCalled(buff, index)
	}
	return nil
}

func (dwm *databaseWriterMock) checkAndCreateIndex(_ string, _ io.Reader) error {
	return nil
}

func newTestElasticSearchDatabase(elasticsearchWriter databaseWriterHandler, arguments elasticSearchDatabaseArgs) *elasticSearchDatabase {
	return &elasticSearchDatabase{
		client:      elasticsearchWriter,
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
	txPool := make(map[string]data.TransactionHandler)

	txPool["tx1"] = &transaction.Transaction{
		Nonce:     uint64(1),
		Value:     big.NewInt(1),
		RcvAddr:   []byte("receiver_address1"),
		SndAddr:   []byte("sender_address1"),
		GasPrice:  uint64(10000),
		GasLimit:  uint64(1000),
		Data:      []byte("tx_data1"),
		Signature: []byte("signature1"),
	}

	txPool["tx2"] = &transaction.Transaction{
		Nonce:     uint64(2),
		Value:     big.NewInt(2),
		RcvAddr:   []byte("receiver_address2"),
		SndAddr:   []byte("sender_address2"),
		GasPrice:  uint64(10000),
		GasLimit:  uint64(1000),
		Data:      []byte("tx_data2"),
		Signature: []byte("signature2"),
	}

	txPool["tx3"] = &transaction.Transaction{
		Nonce:     uint64(3),
		Value:     big.NewInt(3),
		RcvAddr:   []byte("receiver_address3"),
		SndAddr:   []byte("sender_address3"),
		GasPrice:  uint64(10000),
		GasLimit:  uint64(1000),
		Data:      []byte("tx_data3"),
		Signature: []byte("signature3"),
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
	t.Parallel()

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
	t.Parallel()

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	localErr := errors.New("localErr")
	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
			return localErr
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveHeader(header, signerIndexes)

	require.True(t, strings.Contains(output.String(), localErr.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticseachDatabaseSaveHeader_CheckRequestBody(t *testing.T) {
	t.Parallel()

	header := &dataBlock.Header{Nonce: 1}
	signerIndexes := []uint64{0, 1}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
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
	elasticDatabase.saveHeader(header, signerIndexes)
}

func TestElasticseachSaveTransactions(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	localErr := errors.New("localErr")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localErr
		},
	}

	body := newTestBlockBody()
	header := &dataBlock.Header{Nonce: 1, TxCount: 2}
	txPool := newTestTxPool()
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveTransactions(body, header, txPool, 0)
	require.True(t, strings.Contains(output.String(), "indexing bulk of transactions"))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticsearch_saveShardValidatorsPubKeys_RequestError(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	shardId := uint32(0)
	valPubKeys := []string{"key1", "key2"}
	localErr := errors.New("localErr")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
			return localErr
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveShardValidatorsPubKeys(shardId, valPubKeys)

	require.True(t, strings.Contains(output.String(), localErr.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticsearch_saveShardValidatorsPubKeys(t *testing.T) {
	t.Parallel()

	shardId := uint32(0)
	valPubKeys := []string{"key1", "key2"}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(shardId), 10), req.DocumentID)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveShardValidatorsPubKeys(shardId, valPubKeys)
}

func TestElasticsearch_saveShardStatistics_reqError(t *testing.T) {
	t.Parallel()

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
	dbWriter := &databaseWriterMock{
		doBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveShardStatistics(tpsBenchmark)
	require.True(t, strings.Contains(output.String(), localError.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticsearch_saveShardStatistics(t *testing.T) {
	t.Parallel()

	tpsBenchmark := &mock.TpsBenchmarkMock{}
	metaBlock := &dataBlock.MetaBlock{
		TxCount: 2, Nonce: 1,
		ShardInfo: []dataBlock.ShardData{{HeaderHash: []byte("hash")}},
	}
	tpsBenchmark.UpdateWithShardStats(metaBlock)

	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doBulkRequestCalled: func(buff *bytes.Buffer, index string) error {
			require.Equal(t, tpsIndex, index)
			return nil
		},
	}
	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveShardStatistics(tpsBenchmark)
}

func TestElasticsearch_saveRoundInfo(t *testing.T) {
	t.Parallel()

	roundInfo := RoundInfo{
		Index: 1, ShardId: 0, BlockWasProposed: true,
	}
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
			require.Equal(t, strconv.FormatUint(uint64(roundInfo.ShardId), 10)+"_"+strconv.FormatUint(roundInfo.Index, 10), req.DocumentID)
			return nil
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveRoundInfo(roundInfo)
}

func TestElasticsearch_saveRoundInfoRequestError(t *testing.T) {
	t.Parallel()

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})

	roundInfo := RoundInfo{}
	localError := errors.New("local err")
	arguments := createMockElasticsearchDatabaseArgs()
	dbWriter := &databaseWriterMock{
		doRequestCalled: func(req esapi.IndexRequest) error {
			return localError
		},
	}

	elasticDatabase := newTestElasticSearchDatabase(dbWriter, arguments)
	elasticDatabase.saveRoundInfo(roundInfo)
	require.True(t, strings.Contains(output.String(), localError.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}
