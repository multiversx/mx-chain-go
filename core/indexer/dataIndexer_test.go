package indexer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/hashing/sha256"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/testscommon"
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

func NewDataIndexerArguments() DataIndexerArgs {
	return DataIndexerArgs{
		Url:                      "Url",
		UserName:                 "user",
		Password:                 "password",
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		Options:                  &Options{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		EpochStartNotifier:       &mock.EpochStartNotifierStub{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(96),
	}
}

func TestDataIndexer_NewIndexerWithNilUrlShouldError(t *testing.T) {

	arguments := NewDataIndexerArguments()
	arguments.Url = ""
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilUrl, err)
}

func TestDataIndexer_NewIndexerWithNilMarsharlizerShouldError(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestDataIndexer_NewIndexerWithNilHasherShouldError(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.Hasher = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilHasher, err)
}

func TestDataIndexer_NewIndexerWithNilNodesCoordinatorShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.NodesCoordinator = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilNodesCoordinator, err)
}

func TestDataIndexer_NewIndexerWithNilEpochStartNotifierShouldErr(t *testing.T) {
	arguments := NewDataIndexerArguments()
	arguments.EpochStartNotifier = nil
	ei, err := NewDataIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilEpochStartNotifier, err)
}

func TestDataIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blocks" {
			w.WriteHeader(http.StatusOK)
		}
		if r.URL.Path == "/transactions" {
			w.WriteHeader(http.StatusOK)
		}
		if r.URL.Path == "/tps" {
			w.WriteHeader(http.StatusOK)
		}
	}))

	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL
	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
	require.False(t, ei.IsNilIndexer())
}

func TestNewDataIndexerIncorrectUrl(t *testing.T) {
	url := string([]byte{1, 2, 3})

	arguments := NewDataIndexerArguments()
	arguments.Url = url
	ind, err := NewDataIndexer(arguments)
	require.Nil(t, ind)
	require.NotNil(t, err)
}

func TestDataIndexer_UpdateTPS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL

	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.StopIndexing()

	tpsBench := testscommon.TpsBenchmarkMock{}
	tpsBench.Update(newTestMetaBlock())

	ei.UpdateTPS(&tpsBench)
	require.Equal(t, 1, ei.GetQueueLength())
}

func TestDataIndexer_UpdateTPSNil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei, err := NewDataIndexer(arguments)
	require.Nil(t, err)
	_ = ei.StopIndexing()

	ei.UpdateTPS(nil)
	require.NotEmpty(t, output.String())
}

func TestDataIndexer_SaveBlock(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := NewDataIndexer(arguments)
	_ = ei.StopIndexing()

	ei.SaveBlock(&dataBlock.Body{MiniBlocks: []*dataBlock.MiniBlock{}}, nil, nil, nil, nil)
	require.Equal(t, 1, ei.GetQueueLength())
}

func TestDataIndexer_SaveRoundInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	arguments := NewDataIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.Url = ts.URL
	ei, _ := NewDataIndexer(arguments)
	_ = ei.StopIndexing()

	ei.SaveRoundsInfos([]RoundInfo{})
	require.Equal(t, 1, ei.GetQueueLength())
}

func TestDataIndexer_SaveValidatorsPubKeys(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	ei, _ := NewDataIndexer(arguments)

	valPubKey := make(map[uint32][][]byte)

	keys := [][]byte{[]byte("key")}
	valPubKey[0] = keys
	epoch := uint32(0)

	ei.SaveValidatorsPubKeys(valPubKey, epoch)
	require.Equal(t, 1, ei.GetQueueLength())
}

func TestDataIndexer_EpochChange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	getEligibleValidatorsCalled := false

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL
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
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewDataIndexerArguments()
	arguments.Url = ts.URL
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

	dataIndexer, err := NewDataIndexer(DataIndexerArgs{
		Url:            "https://search-elrond-test-okohrj6g5r575cvmkwfv6jraki.eu-west-1.es.amazonaws.com/",
		IndexTemplates: indexTemplates,
		IndexPolicies:  indexPolicies,
		Options: &Options{
			TxIndexingEnabled: true,
		},
		Marshalizer:              &marshal.JsonMarshalizer{},
		Hasher:                   &sha256.Sha256{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		EpochStartNotifier:       &mock.EpochStartNotifierStub{},
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

		go dataIndexer.SaveBlock(body, header, txsPool, signers, []string{"aaaaa", "bbbb"})
	}

	for dataIndexer.GetQueueLength() != 0 {
		time.Sleep(time.Second)
	}
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

func getIndexTemplateAndPolicies() (map[string]io.Reader, map[string]io.Reader) {
	indexTemplates := make(map[string]io.Reader)
	indexPolicies := make(map[string]io.Reader)
	opendistroTemplate, _ := core.OpenFile("./testdata/opendistro.json")
	indexTemplates["opendistro"] = opendistroTemplate
	transactionsTemplate, _ := core.OpenFile("./testdata/transactions.json")
	indexTemplates["transactions"] = transactionsTemplate
	blocksTemplate, _ := core.OpenFile("./testdata/blocks.json")
	indexTemplates["blocks"] = blocksTemplate
	miniblocksTemplate, _ := core.OpenFile("./testdata/miniblocks.json")
	indexTemplates["miniblocks"] = miniblocksTemplate
	transactionsPolicy, _ := core.OpenFile("./testdata/transactions_policy.json")
	indexPolicies["transactions_policy"] = transactionsPolicy
	blocksPolicy, _ := core.OpenFile("./testdata/blocks_policy.json")
	indexPolicies["blocks_policy"] = blocksPolicy

	tpsTemplate, _ := core.OpenFile("./testdata/tps.json")
	indexTemplates["tps"] = tpsTemplate

	return indexTemplates, indexPolicies
}
