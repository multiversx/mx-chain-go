package factory

import (
	"errors"
	"testing"

	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler() *OutportFactoryArgs {
	mockElasticArgs := &indexerFactory.ArgsIndexerFactory{}
	return &OutportFactoryArgs{
		ElasticIndexerFactoryArgs: mockElasticArgs,
	}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() *OutportFactoryArgs
		exError  error
	}{
		{
			name: "NilArgsElasticDriver",
			argsFunc: func() *OutportFactoryArgs {
				return nil
			},
			exError: outport.ErrNilArgsOutportFactory,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *OutportFactoryArgs {
				return createMockArgsOutportHandler()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateOutport(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}
