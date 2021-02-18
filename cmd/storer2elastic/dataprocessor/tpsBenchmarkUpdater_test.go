package dataprocessor

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestNewTPSBenchmarkUpdater(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (genesisNodesConfig sharding.GenesisNodesSetupHandler, tpsIndexer StorageDataIndexer)
		exError  error
	}{
		{
			name: "NilGenesisConfig",
			argsFunc: func() (genesisNodesConfig sharding.GenesisNodesSetupHandler, tpsIndexer StorageDataIndexer) {
				return nil, &mock.ElasticIndexerStub{}
			},
			exError: ErrNilGenesisNodesSetup,
		},
		{
			name: "NilElasticIndexer",
			argsFunc: func() (genesisNodesConfig sharding.GenesisNodesSetupHandler, tpsIndexer StorageDataIndexer) {
				return &mock.GenesisNodesSetupHandlerStub{}, nil
			},
			exError: ErrNilElasticIndexer,
		},
		{
			name: "All arguments ok",
			argsFunc: func() (genesisNodesConfig sharding.GenesisNodesSetupHandler, tpsIndexer StorageDataIndexer) {
				return &mock.GenesisNodesSetupHandlerStub{}, &mock.ElasticIndexerStub{}
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewTPSBenchmarkUpdater(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}
