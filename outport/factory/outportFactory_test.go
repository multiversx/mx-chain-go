package factory_test

import (
	"errors"
	"testing"

	covalentFactory "github.com/ElrondNetwork/covalent-indexer-go/factory"
	indexerFactory "github.com/ElrondNetwork/elastic-indexer-go/factory"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/ElrondNetwork/elrond-go/outport/factory"
	"github.com/ElrondNetwork/elrond-go/process/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon/hashingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	notifierFactory "github.com/ElrondNetwork/notifier-go/factory"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler(indexerEnabled, notifierEnabled, covalentEnabled bool) *factory.OutportFactoryArgs {
	mockElasticArgs := &indexerFactory.ArgsIndexerFactory{
		Enabled: indexerEnabled,
	}
	mockNotifierArgs := &notifierFactory.EventNotifierFactoryArgs{
		Enabled: notifierEnabled,
	}
	mockCovalentArgs := &covalentFactory.ArgsCovalentIndexerFactory{
		Enabled: covalentEnabled,
	}
	return &factory.OutportFactoryArgs{
		ElasticIndexerFactoryArgs:  mockElasticArgs,
		EventNotifierFactoryArgs:   mockNotifierArgs,
		CovalentIndexerFactoryArgs: mockCovalentArgs,
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
			name: "AllOkShouldWork",
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, false, false)
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
				return createMockArgsOutportHandler(true, false, false)
			},
		},
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, true, false)
			},
		},
		{
			argsFunc: func() *factory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, false, true)
			},
		},
	}

	for _, currTest := range tests {
		_, err := factory.CreateOutport(currTest.argsFunc())
		require.NotNil(t, err)
	}
}

func TestCreateOutport_SubscribeCovalentDriver(t *testing.T) {
	args := createMockArgsOutportHandler(false, false, true)

	args.CovalentIndexerFactoryArgs.Hasher = &hashingMocks.HasherMock{}
	args.CovalentIndexerFactoryArgs.ShardCoordinator = &mock.ShardCoordinatorStub{}
	args.CovalentIndexerFactoryArgs.Marshaller = &mock.MarshalizerMock{}
	args.CovalentIndexerFactoryArgs.Accounts = &stateMock.AccountsStub{}
	args.CovalentIndexerFactoryArgs.PubKeyConverter = &mock.PubkeyConverterStub{}

	outPort, err := factory.CreateOutport(args)

	defer func(c outport.OutportHandler) {
		_ = c.Close()
	}(outPort)

	require.True(t, outPort.HasDrivers())
	require.Nil(t, err)
}

func TestCreateOutport_SubscribeNotifierDriver(t *testing.T) {
	args := createMockArgsOutportHandler(false, true, false)

	args.EventNotifierFactoryArgs.Marshalizer = &mock.MarshalizerMock{}
	args.EventNotifierFactoryArgs.Hasher = &hashingMocks.HasherMock{}
	outPort, err := factory.CreateOutport(args)

	defer func(c outport.OutportHandler) {
		_ = c.Close()
	}(outPort)

	require.True(t, outPort.HasDrivers())
	require.Nil(t, err)
}
