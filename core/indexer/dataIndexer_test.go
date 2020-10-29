package indexer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer/disabled"
	"github.com/ElrondNetwork/elrond-go/core/indexer/workItems"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetaBlock() *dataBlock.MetaBlock {
	shardData := dataBlock.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []dataBlock.MiniBlockHeader{},
		TxCount:               100,
	}
	return &dataBlock.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   100,
		ShardInfo: []dataBlock.ShardData{shardData},
	}
}

func NewDataIndexerArguments() ArgDataIndexer {
	return ArgDataIndexer{
		Marshalizer:        &mock.MarshalizerMock{},
		Options:            &Options{},
		NodesCoordinator:   &mock.NodesCoordinatorMock{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		DataDispatcher:     &mock.DispatcherMock{},
		ElasticProcessor:   &mock.ElasticProcessorStub{},
	}
}

func TestDataIndexer_NewIndexerWithNilNodesCoordinatorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.NodesCoordinator = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilNodesCoordinator, err)
}

func TestDataIndexer_NewIndexerWithNilDataDispatcherShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilDataDispatcher, err)
}

func TestDataIndexer_NewIndexerWithNilElasticProcessorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, ErrNilElasticProcessor, err)
}

func TestDataIndexer_NewIndexerWithNilMarshalizerShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestDataIndexer_NewIndexerWithNilEpochStartNotifierShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.EpochStartNotifier = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilEpochStartNotifier, err)
}

func TestDataIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
	arguments := NewDataIndexerArguments()

	ei, err := NewDataIndexer(arguments)

	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
	require.False(t, ei.IsNilIndexer())
}

func TestDataIndexer_UpdateTPS(t *testing.T) {
	t.Parallel()

	called := false
	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.Close()

	tpsBench := testscommon.TpsBenchmarkMock{}
	tpsBench.Update(newTestMetaBlock())

	ei.UpdateTPS(&tpsBench)
	require.True(t, called)
}

func TestDataIndexer_UpdateTPSNil(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.Close()

	ei.UpdateTPS(nil)
	require.NotEmpty(t, output.String())
}

func TestDataIndexer_SaveBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.SaveBlock(&dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{}}, nil,
		nil, nil, nil, []byte("hash"))
	require.True(t, called)
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}

	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	ei, _ := NewDataIndexer(arguments)
	_ = ei.Close()

	ei.SaveRoundsInfo([]workItems.RoundInfo{})
	require.True(t, called)
}

func TestDataIndexer_SaveValidatorsPubKeys(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	valPubKey := make(map[uint32][][]byte)

	keys := [][]byte{[]byte("key")}
	valPubKey[0] = keys
	epoch := uint32(0)

	ei.SaveValidatorsPubKeys(valPubKey, epoch)
	require.True(t, called)
}

func TestDataIndexer_SaveValidatorsRating(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.SaveValidatorsRating("ID", []workItems.ValidatorRatingInfo{
		{Rating: 1}, {Rating: 2},
	})
	require.True(t, called)
}

func TestDataIndexer_RevertIndexedBlock(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.DataDispatcher = &mock.DispatcherMock{
		AddCalled: func(item workItems.WorkItemHandler) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.RevertIndexedBlock(&dataBlock.Header{}, &dataBlock.Body{})
	require.True(t, called)
}

func TestDataIndexer_SetTxLogsProcessor(t *testing.T) {
	called := false

	arguments := NewDataIndexerArguments()
	arguments.ElasticProcessor = &mock.ElasticProcessorStub{
		SetTxLogsProcessorCalled: func(txLogsProc process.TransactionLogProcessorDatabase) {
			called = true
		},
	}
	ei, _ := NewDataIndexer(arguments)

	ei.SetTxLogsProcessor(disabled.NewNilTxLogsProcessor())
	require.True(t, called)
}

func TestDataIndexer_EpochChange(t *testing.T) {
	getEligibleValidatorsCalled := false

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.ShardID = core.MetachainShardId
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup
	wg.Add(1)

	testEpoch := uint32(1)
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()
			if testEpoch == epoch {
				getEligibleValidatorsCalled = true
			}

			return nil, nil
		},
	}

	ei, _ := NewDataIndexer(arguments)
	assert.NotNil(t, ei)

	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: testEpoch})
	wg.Wait()

	assert.True(t, getEligibleValidatorsCalled)
}

func TestDataIndexer_EpochChangeValidators(t *testing.T) {
	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.ShardID = core.MetachainShardId
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup

	val1PubKey := []byte("val1")
	val2PubKey := []byte("val2")
	val1MetaPubKey := []byte("val3")
	val2MetaPubKey := []byte("val4")

	validatorsEpoch1 := map[uint32][][]byte{
		0:                     {val1PubKey, val2PubKey},
		core.MetachainShardId: {val1MetaPubKey, val2MetaPubKey},
	}
	validatorsEpoch2 := map[uint32][][]byte{
		0:                     {val2PubKey, val1PubKey},
		core.MetachainShardId: {val2MetaPubKey, val1MetaPubKey},
	}
	var firstEpochCalled, secondEpochCalled bool
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()

			switch epoch {
			case 1:
				firstEpochCalled = true
				return validatorsEpoch1, nil
			case 2:
				secondEpochCalled = true
				return validatorsEpoch2, nil
			default:
				return nil, nil
			}
		},
	}

	ei, _ := NewDataIndexer(arguments)
	assert.NotNil(t, ei)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: 1})
	wg.Wait()
	assert.True(t, firstEpochCalled)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 10, Epoch: 2})
	wg.Wait()
	assert.True(t, secondEpochCalled)
}

func TestDataIndexer(t *testing.T) {
	t.Skip("this is not a short test")

	testCreateIndexer(t)
}

func testCreateIndexer(t *testing.T) {
	indexTemplates, indexPolicies := getIndexTemplateAndPolicies()

	dispatcher, _ := NewDataDispatcher(100)
	dbClient, _ := NewElasticClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Username:  "",
		Password:  "",
	})

	elasticIndexer, _ := NewElasticProcessor(ArgElasticProcessor{
		IndexTemplates:           indexTemplates,
		IndexPolicies:            indexPolicies,
		Marshalizer:              &marshal.JsonMarshalizer{},
		Hasher:                   &sha256.Sha256{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		Options:                  &Options{},
		DBClient:                 dbClient,
	})

	di, err := NewDataIndexer(ArgDataIndexer{
		Options:            &Options{},
		Marshalizer:        &marshal.JsonMarshalizer{},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		DataDispatcher:     dispatcher,
		ElasticProcessor:   elasticIndexer,
	})
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	// Generate transaction and hashes
	numTransactions := 10
	dataSize := 1000
	signers := []uint64{395, 207, 16, 99, 358, 292, 258, 362, 161, 247, 1, 137, 91, 309, 30, 92, 166, 361, 158, 301, 218, 80, 108, 392, 153, 343, 110, 133, 351, 316, 5, 305, 248, 123,
		327, 322, 97, 86, 215, 212, 289, 250, 229, 13, 237, 20, 269, 37, 243, 29, 236, 155, 338, 257, 375, 142, 129, 93, 234, 195, 377, 311, 170}
	for i := 0; i < 100; i++ {
		txs, hashes := generateTransactions(numTransactions, dataSize)

		header := &dataBlock.Header{
			Nonce: uint64(i),
		}
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

		di.SaveBlock(body, header, txsPool, signers, []string{"aaaaa", "bbbb"}, []byte("hash"))
	}

	time.Sleep(100 * time.Second)
}

func generateTransactions(numTxs int, datFieldSize int) ([]transaction.Transaction, []string) {
	txs := make([]transaction.Transaction, numTxs)
	hashes := make([]string, numTxs)

	randomByteArray := make([]byte, datFieldSize)
	_, _ = rand.Read(randomByteArray)

	for i := 0; i < numTxs; i++ {
		txs[i] = transaction.Transaction{
			Nonce:     uint64(i),
			Value:     big.NewInt(int64(i)),
			RcvAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			SndAddr:   []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
			GasPrice:  200000000000,
			GasLimit:  20000,
			Data:      randomByteArray,
			Signature: []byte("443e79a8d99ba093262c1db48c58ab3d59bcfeb313ca5cddf2a9d1d06f9894ec"),
		}
		hashes[i] = fmt.Sprintf("%v", time.Now())
	}

	return txs, hashes
}

func getIndexTemplateAndPolicies() (map[string]*bytes.Buffer, map[string]*bytes.Buffer) {
	indexTemplates := make(map[string]*bytes.Buffer)
	indexPolicies := make(map[string]*bytes.Buffer)

	template := &bytes.Buffer{}
	_ = core.LoadJsonFile(template, "./testdata/opendistro.json")
	indexTemplates["opendistro"] = template
	_ = core.LoadJsonFile(template, "./testdata/transactions.json")
	indexTemplates["transactions"] = template

	_ = core.LoadJsonFile(template, "./testdata/blocks.json")
	indexTemplates["blocks"] = template
	_ = core.LoadJsonFile(template, "./testdata/miniblocks.json")
	indexTemplates["miniblocks"] = template

	_ = core.LoadJsonFile(template, "./testdata/tps.json")
	indexTemplates["tps"] = template

	policy := &bytes.Buffer{}
	_ = core.LoadJsonFile(template, "./testdata/transactions_policy.json")
	indexPolicies["transactions_policy"] = policy
	_ = core.LoadJsonFile(template, "./testdata/blocks_policy.json")
	indexPolicies["blocks_policy"] = policy

	return indexTemplates, indexPolicies
}
