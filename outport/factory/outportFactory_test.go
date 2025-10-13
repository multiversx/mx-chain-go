package factory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	indexerFactory "github.com/multiversx/mx-chain-es-indexer-go/process/factory"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/outport"
	outportFactory "github.com/multiversx/mx-chain-go/outport/factory"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler(indexerEnabled, notifierEnabled bool) *outportFactory.OutportFactoryArgs {
	mockElasticArgs := indexerFactory.ArgsIndexerFactory{
		Enabled: indexerEnabled,
	}
	mockNotifierArgs := &outportFactory.EventNotifierFactoryArgs{
		Enabled: notifierEnabled,
	}
	return &outportFactory.OutportFactoryArgs{
		RetrialInterval:           time.Second,
		ElasticIndexerFactoryArgs: mockElasticArgs,
		EventNotifierFactoryArgs:  mockNotifierArgs,
		EnableEpochsHandler:       &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler:       &testscommon.EnableRoundsHandlerStub{},
	}
}

func TestNewIndexerFactory(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		argsFunc func() *outportFactory.OutportFactoryArgs
		exError  error
	}{
		{
			name: "NilArgsOutportFactory",
			argsFunc: func() *outportFactory.OutportFactoryArgs {
				return nil
			},
			exError: outport.ErrNilArgsOutportFactory,
		},
		{
			name: "invalid retrial duration",
			argsFunc: func() *outportFactory.OutportFactoryArgs {
				args := createMockArgsOutportHandler(false, false)
				args.RetrialInterval = 0
				return args
			},
			exError: outport.ErrInvalidRetrialInterval,
		},
		{
			name: "AllOkShouldWork",
			argsFunc: func() *outportFactory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, false)
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := outportFactory.CreateOutport(tt.argsFunc())
			require.True(t, errors.Is(err, tt.exError))
		})
	}
}

func TestCreateOutport_EnabledDriversNilMockArgsExpectErrorSubscribingDrivers(t *testing.T) {
	t.Parallel()
	tests := []struct {
		argsFunc func() *outportFactory.OutportFactoryArgs
	}{
		{
			argsFunc: func() *outportFactory.OutportFactoryArgs {
				return createMockArgsOutportHandler(true, false)
			},
		},
		{
			argsFunc: func() *outportFactory.OutportFactoryArgs {
				return createMockArgsOutportHandler(false, true)
			},
		},
	}

	for _, currTest := range tests {
		_, err := outportFactory.CreateOutport(currTest.argsFunc())
		require.NotNil(t, err)
	}
}

func TestCreateOutport_SubscribeNotifierDriver(t *testing.T) {
	args := createMockArgsOutportHandler(false, true)

	args.EventNotifierFactoryArgs.Marshaller = &mock.MarshalizerMock{}
	args.EventNotifierFactoryArgs.RequestTimeoutSec = 1
	outPort, err := outportFactory.CreateOutport(args)
	require.Nil(t, err)

	defer func(c outport.OutportHandler) {
		_ = c.Close()
	}(outPort)

	require.True(t, outPort.HasDrivers())
}

func TestCreateOutport_SubscribeMultipleHostDrivers(t *testing.T) {
	args := &outportFactory.OutportFactoryArgs{
		RetrialInterval: time.Second,
		EventNotifierFactoryArgs: &outportFactory.EventNotifierFactoryArgs{
			Enabled: false,
		},
		ElasticIndexerFactoryArgs: indexerFactory.ArgsIndexerFactory{
			Enabled: false,
		},
		HostDriversArgs: []outportFactory.ArgsHostDriverFactory{
			{
				Marshaller: &marshallerMock.MarshalizerMock{},
				HostConfig: config.HostDriversConfig{
					Enabled:            true,
					URL:                "ws://localhost",
					RetryDurationInSec: 1,
					MarshallerType:     "json",
					Mode:               data.ModeClient,
				},
			},
			{
				Marshaller: &marshallerMock.MarshalizerMock{},
				HostConfig: config.HostDriversConfig{
					Enabled:            false,
					URL:                "ws://localhost",
					RetryDurationInSec: 1,
					MarshallerType:     "json",
					Mode:               data.ModeClient,
				},
			},
			{
				Marshaller: &marshallerMock.MarshalizerMock{},
				HostConfig: config.HostDriversConfig{
					Enabled:            true,
					URL:                "ws://localhost",
					RetryDurationInSec: 1,
					MarshallerType:     "json",
					Mode:               data.ModeClient,
				},
			},
		},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
	}

	outPort, err := outportFactory.CreateOutport(args)
	require.Nil(t, err)

	defer func() {
		_ = outPort.Close()
	}()

	require.True(t, outPort.HasDrivers())
}

func TestCreateAndSubscribeDriversShouldReturnError(t *testing.T) {
	args := &outportFactory.OutportFactoryArgs{
		RetrialInterval: time.Second,
		EventNotifierFactoryArgs: &outportFactory.EventNotifierFactoryArgs{
			Enabled: false,
		},
		ElasticIndexerFactoryArgs: indexerFactory.ArgsIndexerFactory{
			Enabled: false,
		},
		HostDriversArgs: []outportFactory.ArgsHostDriverFactory{
			{
				Marshaller: &marshallerMock.MarshalizerMock{},
				HostConfig: config.HostDriversConfig{
					Enabled:            true,
					URL:                "localhost",
					RetryDurationInSec: 1,
					MarshallerType:     "json",
					Mode:               "wrong mode",
				},
			},
		},
		EnableEpochsHandler: &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EnableRoundsHandler: &testscommon.EnableRoundsHandlerStub{},
	}

	outPort, err := outportFactory.CreateOutport(args)
	require.Nil(t, outPort)
	require.ErrorIs(t, err, data.ErrInvalidWebSocketHostMode)
}
