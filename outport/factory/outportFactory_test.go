package factory_test

import (
	"errors"
	"testing"
	"time"

	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/factory"
	notifierFactory "github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler(indexerEnabled, notifierEnabled bool) *factory.OutportFactoryArgs {
	mockElasticArgs := indexerFactory.ArgsIndexerFactory{
		Enabled: indexerEnabled,
	}
	mockNotifierArgs := &notifierFactory.EventNotifierFactoryArgs{
		Enabled: notifierEnabled,
	}
	return &factory.OutportFactoryArgs{
		RetrialInterval:           time.Second,
		ElasticIndexerFactoryArgs: mockElasticArgs,
		EventNotifierFactoryArgs:  mockNotifierArgs,
	}
}

func TestNewIndexerFactory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		argsFunc func() *factory.OutportFactoryArgs
		exError  error
	}{
		{
			name: "NilArgsOutportFactory",
			argsFunc: func() *factory.OutportFactoryArgs {
				return nil
			},
			exError: outport.ErrNilArgsOutportFactory,
		},
		{
			name: "invalid retrial duration",
			argsFunc: func() *factory.OutportFactoryArgs {
				args := createMockArgsOutportHandler(false, false)
				args.RetrialInterval = 0
				return args
			},
			exError: outport.ErrInvalidRetrialInterval,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, false)
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := factory.CreateOutport(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestCreateOutport_EnabledDriversNilMockArgsExpectErrorSubscribingDrivers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argsFunc func() *factory.OutportFactoryArgs
	}{
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(true, false)
			},
		},
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, true)
			},
		},
	}

	for _, currTest := range tests {
		_, err := factory.CreateOutport(currTest.argsFunc())
		require.NotNil(t, err)
	}
}

func TestCreateOutport_SubscribeNotifierDriver(t *testing.T) {
	args := createMockArgsOutportHandler(false, true)

	args.EventNotifierFactoryArgs.Marshaller = &mock.MarshalizerMock{}
	args.EventNotifierFactoryArgs.RequestTimeoutSec = 1
	outPort, err := factory.CreateOutport(args)
	require.Nil(t, err)

	defer func(c outport.OutportHandler) {
		_ = c.Close()
	}(outPort)

	require.True(t, outPort.HasDrivers())
}
