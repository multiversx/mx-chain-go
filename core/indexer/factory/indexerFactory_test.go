package factory

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/core/mock"
	"github.com/stretchr/testify/require"
)

func createMockIndexerFactoryArgs() *ArgsIndexerFactory {
	return &ArgsIndexerFactory{
		Enabled:                  true,
		IndexerCacheSize:         100,
		ShardID:                  0,
		Url:                      "test-url",
		UserName:                 "",
		Password:                 "",
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		EpochStartNotifier:       &mock.EpochStartNotifierStub{},
		NodesCoordinator:         &mock.NodesCoordinatorMock{},
		AddressPubkeyConverter:   &mock.PubkeyConverterMock{},
		ValidatorPubkeyConverter: &mock.PubkeyConverterMock{},
		IndexTemplates:           nil,
		IndexPolicies:            nil,
		Options:                  &indexer.Options{},
	}
}

func TestNewIndexerFactory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() *ArgsIndexerFactory
		exError  error
	}{
		{
			name: "InvalidCacheSize",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.IndexerCacheSize = 0
				return args
			},
			exError: indexer.ErrInvalidCacheSize,
		},
		{
			name: "NilAddressPubkeyConverter",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.AddressPubkeyConverter = nil
				return args
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilValidatorPubkeyConverter",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.ValidatorPubkeyConverter = nil
				return args
			},
			exError: indexer.ErrNilPubkeyConverter,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Marshalizer = nil
				return args
			},
			exError: core.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Hasher = nil
				return args
			},
			exError: core.ErrNilHasher,
		},
		{
			name: "NilNodesCoordinator",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.NodesCoordinator = nil
				return args
			},
			exError: core.ErrNilNodesCoordinator,
		},
		{
			name: "NilEpochStartNotifier",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.EpochStartNotifier = nil
				return args
			},
			exError: core.ErrNilEpochStartNotifier,
		},
		{
			name: "EmptyUrl",
			argsFunc: func() *ArgsIndexerFactory {
				args := createMockIndexerFactoryArgs()
				args.Url = ""
				return args
			},
			exError: core.ErrNilUrl,
		},
		{
			name: "All arguments ok",
			argsFunc: func() *ArgsIndexerFactory {
				return createMockIndexerFactoryArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewIndexerFactory(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestIndexerFactoryCreate_NilIndexer(t *testing.T) {
	t.Parallel()

	args := createMockIndexerFactoryArgs()
	args.Enabled = false
	indexerFactory, err := NewIndexerFactory(args)
	require.NoError(t, err)

	nilIndexer, err := indexerFactory.Create()
	require.NoError(t, err)

	_, ok := nilIndexer.(*indexer.NilIndexer)
	require.True(t, ok)
}

func TestIndexerFactoryCreate_ElasticIndexer(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	args := createMockIndexerFactoryArgs()
	args.Url = ts.URL

	indexerFactory, err := NewIndexerFactory(args)
	require.NoError(t, err)

	elasticIndexer, err := indexerFactory.Create()
	require.NoError(t, err)

	err = elasticIndexer.StopIndexing()
	require.NoError(t, err)
	require.False(t, elasticIndexer.IsNilIndexer())

}
