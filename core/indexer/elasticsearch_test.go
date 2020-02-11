package indexer_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/logger"
	"github.com/stretchr/testify/require"
)

func newTestMetaBlock() *block.MetaBlock {
	shardData := block.ShardData{
		ShardID:               1,
		HeaderHash:            []byte{1},
		ShardMiniBlockHeaders: []block.ShardMiniBlockHeader{},
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
		Url:              "url",
		UserName:         "user",
		Password:         "password",
		ShardCoordinator: &mock.ShardCoordinatorMock{},
		Marshalizer:      &mock.MarshalizerMock{},
		Hasher:           &mock.HasherMock{},
		Options:          &indexer.Options{},
	}
}

func TestElasticIndexer_NewIndexerWithNilUrlShouldError(t *testing.T) {

	arguments := NewElasticIndexerArguments()
	arguments.Url = ""
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilUrl, err)
}

func TestElasticIndexer_NewIndexerWithNilShardCoordinatorShouldError(t *testing.T) {
	arguments := NewElasticIndexerArguments()
	arguments.ShardCoordinator = nil
	ei, err := indexer.NewElasticIndexer(arguments)

	require.Nil(t, ei)
	require.Equal(t, core.ErrNilCoordinator, err)
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
	ei, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, err)

	tpsBench := mock.TpsBenchmarkMock{}
	tpsBench.Update(newTestMetaBlock())

	ei.UpdateTPS(&tpsBench)
	require.Empty(t, output.String())
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticIndexer_UpdateTPSNil(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, err := indexer.NewElasticIndexer(arguments)
	require.Nil(t, err)

	ei.UpdateTPS(nil)
	require.NotEmpty(t, output.String())
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticIndexer_SaveBlockNilHeaderHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	ei.SaveBlock(block.Body{{}}, nil, nil, nil)
	require.True(t, strings.Contains(output.String(), indexer.ErrNoHeader.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticIndexer_SaveBlockNilBodyHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	ei.SaveBlock(nil, nil, nil, nil)
	require.True(t, strings.Contains(output.String(), indexer.ErrBodyTypeAssertion.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticIndexer_SaveBlockNoMiniBlocks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	header := &block.Header{}
	body := block.Body{}
	ei.SaveBlock(body, header, nil, []uint64{0})
	require.True(t, strings.Contains(output.String(), indexer.ErrNoMiniblocks.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}

func TestElasticIndexer_SaveMetaBlockNilHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	output := &bytes.Buffer{}
	_ = logger.SetLogLevel("core/indexer:TRACE")
	_ = logger.AddLogObserver(output, &logger.PlainFormatter{})
	arguments := NewElasticIndexerArguments()
	arguments.Url = ts.URL
	ei, _ := indexer.NewElasticIndexer(arguments)

	ei.SaveMetaBlock(nil, []uint64{0})
	require.True(t, strings.Contains(output.String(), indexer.ErrNoHeader.Error()))
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
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

	ei.SaveRoundInfo(indexer.RoundInfo{})
	require.NotEmpty(t, output.String())
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
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
	ei.SaveValidatorsPubKeys(valPubKey)

	require.Empty(t, output.String())
	_ = logger.RemoveLogObserver(output)
	_ = logger.SetLogLevel("core/indexer:INFO")
}
