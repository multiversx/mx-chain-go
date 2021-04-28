package factory

import (
	"errors"
	"testing"

	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler() *ArgsOutportFactory {
	return &ArgsOutportFactory{
		ArgsElasticDriver: &indexerFactory.ArgsIndexerFactory{},
	}
}

func TestNewIndexerFactory(t *testing.T) {
	tests := []struct {
		name     string
		argsFunc func() *ArgsOutportFactory
		exError  error
	}{
		{
			name: "NilArgsOutportHandler",
			argsFunc: func() *ArgsOutportFactory {
				return nil
			},
			exError: outport.ErrNilArgsOutportFactory,
		},
		{
			name: "NilArgsElasticDriver",
			argsFunc: func() *ArgsOutportFactory {
				args := createMockArgsOutportHandler()
				args.ArgsElasticDriver = nil
				return args
			},
			exError: outport.ErrNilArgsElasticDriverFactory,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *ArgsOutportFactory {
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
