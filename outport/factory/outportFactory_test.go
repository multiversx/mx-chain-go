package factory

import (
	"errors"
	"testing"

	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler() *indexerFactory.ArgsIndexerFactory {
	return &indexerFactory.ArgsIndexerFactory{}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() *indexerFactory.ArgsIndexerFactory
		exError  error
	}{
		{
			name: "NilArgsElasticDriver",
			argsFunc: func() *indexerFactory.ArgsIndexerFactory {
				args := createMockArgsOutportHandler()
				args = nil
				return args
			},
			exError: outport.ErrNilArgsElasticDriverFactory,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *indexerFactory.ArgsIndexerFactory {
				args := createMockArgsOutportHandler()
				return args
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
