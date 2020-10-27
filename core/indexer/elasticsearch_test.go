package indexer_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMetaBlock() *block.MetaBlock {
	shardData := block.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []block.MiniBlockHeader{},
		TxCount:               100,
	}
	return &block.MetaBlock{
		Nonce:     1,
		Round:     2,
		TxCount:   100,
		ShardInfo: []block.ShardData{shardData},
	}
}

func NewElasticIndexerArguments() indexer.ElasticIndexerArgs {
	return indexer.ElasticIndexerArgs{
		Url:                      "Url",
		UserName:                 "user",
		Password:                 "password",
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		Options:                  &indexer.Options{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		EpochStartNotifier:       &mock.EpochStartNotifierStub{},
		AddressPubkeyConverter:   mock.NewPubkeyConverterMock(32),
		ValidatorPubkeyConverter: mock.NewPubkeyConverterMock(96),
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		IsInImportDBMode:         false,
	}
}

func TestElasticIndexer_NewIndexerWithNilUrlShouldError(t *testing.T) {

	arguments := NewElasticIndexerArguments()
	arguments.Url = ""
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilUrl, err)
}

func TestElasticIndexer_NewIndexerWithNilMarsharlizerShouldError(t *testing.T) {
	arguments := NewElasticIndexerArguments()
	arguments.Marshalizer = nil
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilMarshalizer, err)
}

func TestElasticIndexer_NewIndexerWithNilHasherShouldError(t *testing.T) {
	arguments := NewElasticIndexerArguments()
	arguments.Hasher = nil
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilHasher, err)
}

func TestElasticIndexer_NewIndexerWithNilNodesCoordinatorShouldErr(t *testing.T) {
	arguments := NewElasticIndexerArguments()
	arguments.NodesCoordinator = nil
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilNodesCoordinator, err)
}

func TestElasticIndexer_NewIndexerWithNilEpochStartNotifierShouldErr(t *testing.T) {
	arguments := NewElasticIndexerArguments()
	arguments.EpochStartNotifier = nil
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilEpochStartNotifier, err)
}

func TestElasticIndexer_NewIndexerWithCorrectParamsShouldWork(t *testing.T) {
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

	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, err)
	require.False(t, check.IfNil(ei))
	require.False(t, ei.IsNilIndexer())
}

func TestNewElasticIndexerIncorrectUrl(t *testing.T) {
	url := string([]byte{1, 2, 3})

	arguments := NewElasticIndexerArguments()
	arguments.Url = url
	ind, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, ind)
	require.NotNil(t, err)
}

func TestElasticIndexer_UpdateTPS(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, err)

	tpsBench := testscommon.TpsBenchmarkMock{}
	tpsBench.Update(newTestMetaBlock())

	ei.UpdateTPS(&tpsBench)
	require.Empty(t, output.String())
}

func TestElasticIndexer_UpdateTPSNil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, err)

	ei.UpdateTPS(nil)
	require.NotEmpty(t, output.String())
}

func TestElasticIndexer_SaveBlockNilHeaderHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei.SaveBlock(&block.Body{MiniBlocks: []*block.MiniBlock{}}, nil, nil, nil, nil)
	require.True(t, strings.Contains(output.String(), indexer.ErrNoHeader.Error()))
}

func TestElasticIndexer_SaveBlockNilBodyHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei.SaveBlock(nil, nil, nil, nil, nil)
	require.True(t, strings.Contains(output.String(), indexer.ErrBodyTypeAssertion.Error()))
}

func TestElasticIndexer_SaveRoundInfo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	ei.SaveRoundsInfos([]indexer.RoundInfo{})
	require.Empty(t, output.String())
}

func TestElasticIndexer_SaveValidatorsPubKeys(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	ei, _ := indexer.NewElasticIndexer(arguments)

	valPubKey := make(map[uint32][][]byte)

	keys := [][]byte{[]byte("key")}
	valPubKey[0] = keys
	epoch := uint32(0)
	ei.SaveValidatorsPubKeys(valPubKey, epoch)

	defer func() {
		_ = logger.RemoveLogObserver(output)
		_ = logger.SetLogLevel("core/indexer:INFO")
	}()

	require.Empty(t, output.String())
}

func TestElasticIndexer_EpochChange(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	getEligibleValidatorsCalled := false

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	scm := &mock.ShardCoordinatorMock{}
	scm.SetSelfId(core.MetachainShardId)
	arguments.ShardCoordinator = scm
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

	ei, _ := indexer.NewElasticIndexer(arguments)
	assert.NotNil(t, ei)

	epochChangeNotifier.NotifyAll(&block.Header{Nonce: 1, Epoch: testEpoch})
	wg.Wait()

	assert.True(t, getEligibleValidatorsCalled)
}

func TestElasticIndexer_EpochChangeValidators(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	arguments.Marshalizer = &mock.MarshalizerMock{Fail: true}
	scm := &mock.ShardCoordinatorMock{}
	scm.SetSelfId(core.MetachainShardId)
	arguments.ShardCoordinator = scm
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

	ei, _ := indexer.NewElasticIndexer(arguments)
	assert.NotNil(t, ei)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&block.Header{Nonce: 1, Epoch: 1})
	wg.Wait()
	assert.True(t, firstEpochCalled)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&block.Header{Nonce: 10, Epoch: 2})
	wg.Wait()
	assert.True(t, secondEpochCalled)
}
