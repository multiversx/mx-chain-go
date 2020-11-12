package factory

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/ElrondNetwork/elrond-go/outport/drivers/elastic"
	"github.com/stretchr/testify/require"
)

func createMockIndexerFactoryArgs() *ArgsElasticDriverFactory {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	return &ArgsElasticDriverFactory{
		Enabled:                  true,
		IndexerCacheSize:         100,
		Url:                      ts.URL,
		UserName:                 "",
		Password:                 "",
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		TemplatesPath:            "../testdata",
		Options:                  &elastic.Options{},
		EnabledIndexes:           []string{"blocks", "transactions", "miniblocks", "tps", "validators", "round", "accounts", "rating"},
		AccountsDB:               &mock.AccountsStub{},
		FeeConfig:                &config.FeeSettings{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		IsInImportDBMode:         false,
	}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() *ArgsElasticDriverFactory
		exError  error
	}{
		{
			name: "InvalidCacheSize",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.IndexerCacheSize = -1
				return args
			},
			exError: elastic.ErrNegativeCacheSize,
		},
		{
			name: "NilAddressPubkeyConverter",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.AddressPubkeyConverter = nil
				return args
			},
			exError: elastic.ErrNilPubkeyConverter,
		},
		{
			name: "NilValidatorPubkeyConverter",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.ValidatorPubkeyConverter = nil
				return args
			},
			exError: elastic.ErrNilPubkeyConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.Marshalizer = nil
				return args
			},
			exError: core.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.Hasher = nil
				return args
			},
			exError: core.ErrNilHasher,
		},
		{
			name: "NilAccountsDB",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.AccountsDB = nil
				return args
			},
			exError: elastic.ErrNilAccountsDB,
		},
		{
			name: "EmptyUrl",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.Url = ""
				return args
			},
			exError: core.ErrNilUrl,
		},
		{
			name: "NilFeeConfig",
			argsFunc: func() *ArgsElasticDriverFactory {
				args := createMockIndexerFactoryArgs()
				args.FeeConfig = nil
				return args
			},
			exError: core.ErrNilFeeConfig,
		},
		{
			name: "All arguments ok",
			argsFunc: func() *ArgsElasticDriverFactory {
				return createMockIndexerFactoryArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewElasticClient(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestIndexerFactoryCreate_ElasticIndexer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	args := createMockIndexerFactoryArgs()
	args.Url = ts.URL

	elasticIndexer, err := NewElasticClient(args)
	require.NoError(t, err)

	err = elasticIndexer.Close()
	require.NoError(t, err)

	err = elasticIndexer.Close()
	require.NoError(t, err)
}
