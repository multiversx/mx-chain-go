package factory

import (
	"errors"
	"sync"
	"testing"

	logger "github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core"
	dataBlock "github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/outport"
	driversFacotry "github.com/ElrondNetwork/elrond-go/outport/drivers/factory"
	"github.com/ElrondNetwork/elrond-go/outport/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgsOutportHandler() *ArgsOutportFactory {
	return &ArgsOutportFactory{
		ArgsElasticDriver: &driversFacotry.ArgsElasticDriverFactory{
			ShardCoordinator: &mock.ShardCoordinatorMock{},
		},
		EpochStartNotifier: &mock.EpochStartNotifierStub{},
		NodesCoordinator:   &mock.NodesCoordinatorMock{},
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
			name: "NilShardCoordinator",
			argsFunc: func() *ArgsOutportFactory {
				args := createMockArgsOutportHandler()
				args.ArgsElasticDriver.ShardCoordinator = nil
				return args
			},
			exError: outport.ErrNilShardCoordinator,
		},

		{
			name: "NilEpochStartNotifier",
			argsFunc: func() *ArgsOutportFactory {
				args := createMockArgsOutportHandler()
				args.EpochStartNotifier = nil
				return args
			},
			exError: outport.ErrNilEpochStartNotifier,
		},
		{
			name: "NilNodesCoordinator",
			argsFunc: func() *ArgsOutportFactory {
				args := createMockArgsOutportHandler()
				args.NodesCoordinator = nil
				return args
			},
			exError: outport.ErrNilNodesCoordinator,
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

func TestOutportFactory_EpochStartEventHandler(t *testing.T) {
	getEligibleValidatorsCalled := false

	_ = logger.SetLogLevel("core/indexer:TRACE")
	arguments := &ArgsOutportFactory{
		ArgsElasticDriver:  &driversFacotry.ArgsElasticDriverFactory{},
		EpochStartNotifier: nil,
		NodesCoordinator:   nil,
	}
	arguments.ArgsElasticDriver.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup
	wg.Add(1)

	testEpoch := uint32(1)
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()
			if testEpoch == epoch {
				getEligibleValidatorsCalled = true
			}

			return nil, nil
		},
	}

	outportHandler, _ := CreateOutport(arguments)
	assert.NotNil(t, outportHandler)

	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: testEpoch})
	wg.Wait()

	assert.True(t, getEligibleValidatorsCalled)
}

func TestOutportFactory_EpochChangeValidators(t *testing.T) {
	_ = logger.SetLogLevel("core/indexer:TRACE")
	arguments := &ArgsOutportFactory{
		ArgsElasticDriver:  &driversFacotry.ArgsElasticDriverFactory{},
		EpochStartNotifier: nil,
		NodesCoordinator:   nil,
	}
	arguments.ArgsElasticDriver.ShardCoordinator = &mock.ShardCoordinatorMock{
		SelfID: core.MetachainShardId,
	}
	epochChangeNotifier := &mock.EpochStartNotifierStub{}
	arguments.EpochStartNotifier = epochChangeNotifier

	var wg sync.WaitGroup

	val1PubKey := []byte("val1")
	val2PubKey := []byte("val2")
	val1MetaPubKey := []byte("val3")
	val2MetaPubKey := []byte("val4")

	validatorsEpoch1 := map[uint32][][]byte{
		0:                     {val1PubKey, val2PubKey},
		core.MetachainShardId: {val1MetaPubKey, val2MetaPubKey},
	}
	validatorsEpoch2 := map[uint32][][]byte{
		0:                     {val2PubKey, val1PubKey},
		core.MetachainShardId: {val2MetaPubKey, val1MetaPubKey},
	}
	var firstEpochCalled, secondEpochCalled bool
	arguments.NodesCoordinator = &mock.NodesCoordinatorMock{
		GetAllEligibleValidatorsPublicKeysCalled: func(epoch uint32) (m map[uint32][][]byte, err error) {
			defer wg.Done()

			switch epoch {
			case 1:
				firstEpochCalled = true
				return validatorsEpoch1, nil
			case 2:
				secondEpochCalled = true
				return validatorsEpoch2, nil
			default:
				return nil, nil
			}
		},
	}

	outportHandler, _ := CreateOutport(arguments)
	assert.NotNil(t, outportHandler)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 1, Epoch: 1})
	wg.Wait()
	assert.True(t, firstEpochCalled)

	wg.Add(1)
	epochChangeNotifier.NotifyAll(&dataBlock.Header{Nonce: 10, Epoch: 2})
	wg.Wait()
	assert.True(t, secondEpochCalled)
}
