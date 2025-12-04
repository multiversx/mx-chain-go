package p2p_test

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	retriever "github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/epochStart"
	chainP2P "github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/p2p"
	processP2PMocks "github.com/multiversx/mx-chain-go/process/p2p/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

func createArgs() p2p.ArgEpochChangeTopicsHandler {
	return p2p.ArgEpochChangeTopicsHandler{
		MainMessenger:        &p2pmocks.MessengerStub{},
		FullArchiveMessenger: &p2pmocks.MessengerStub{},
		EnableEpochsHandler:  &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
		EpochNotifier: &epochNotifier.EpochNotifierStub{
			RegisterHandlerCalled: func(handler epochStart.ActionHandler) {
			},
		},
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
		ResolversContainer: &dataRetriever.ResolversContainerStub{
			GetCalled: func(key string) (retriever.Resolver, error) {
				return &processP2PMocks.TopicResolverMock{}, nil
			},
		},
		MainInterceptorsContainer:        &testscommon.InterceptorsContainerStub{},
		FullArchiveInterceptorsContainer: &testscommon.InterceptorsContainerStub{},
		IsFullArchive:                    false,
	}
}

func TestNewEpochChangeTopicsHandler(t *testing.T) {
	t.Parallel()

	t.Run("nil main messenger, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.MainMessenger = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, fmt.Errorf("%w for main messenger", process.ErrNilMessenger), err)
	})
	t.Run("nil full archive messenger, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.FullArchiveMessenger = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, fmt.Errorf("%w for full archive messenger", process.ErrNilMessenger), err)
	})
	t.Run("nil enable epochs handler, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.EnableEpochsHandler = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, process.ErrNilEnableEpochsHandler, err)
	})
	t.Run("nil shard coordinator, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.ShardCoordinator = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, process.ErrNilShardCoordinator, err)
	})
	t.Run("nil resolvers container, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.ResolversContainer = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, process.ErrNilResolversContainer, err)
	})
	t.Run("nil main interceptors container, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.MainInterceptorsContainer = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, fmt.Errorf("%w for main container", process.ErrNilInterceptorsContainer), err)
	})
	t.Run("nil full archive interceptors container, should return error", func(t *testing.T) {
		t.Parallel()

		args := createArgs()
		args.FullArchiveInterceptorsContainer = nil
		handler, err := p2p.NewEpochChangeTopicsHandler(args)
		require.Nil(t, handler)
		require.Equal(t, fmt.Errorf("%w for full archive container", process.ErrNilInterceptorsContainer), err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handler, err := p2p.NewEpochChangeTopicsHandler(createArgs())
		require.NotNil(t, handler)
		require.Nil(t, err)
	})
}

func TestEpochChangeTopicsHandler_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createArgs()
	args.EnableEpochsHandler = nil
	handler, _ := p2p.NewEpochChangeTopicsHandler(args)
	require.True(t, handler.IsInterfaceNil())

	handler, _ = p2p.NewEpochChangeTopicsHandler(createArgs())
	require.False(t, handler.IsInterfaceNil())
}

func TestEpochChangeTopicsHandler_EpochConfirmed(t *testing.T) {
	t.Parallel()

	t.Run("topics already moved, should do nothing", func(t *testing.T) {
		t.Parallel()

		handler, _ := p2p.NewEpochChangeTopicsHandler(createArgs())
		require.NotNil(t, handler)

		handler.SetTopicsMoved(true)
		handler.EpochConfirmed(0, 0)
	})
	t.Run("epoch before supernova, should not move topics", func(t *testing.T) {
		t.Parallel()

		providedEpoch := uint32(100)
		args := createArgs()
		args.EnableEpochsHandler = &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			GetActivationEpochCalled: func(flag core.EnableEpochFlag) uint32 {
				return providedEpoch
			},
		}
		handler, _ := p2p.NewEpochChangeTopicsHandler(args)
		require.NotNil(t, handler)

		handler.EpochConfirmed(providedEpoch-1, 0)
	})
	t.Run("should work and replace all topics, shard", func(t *testing.T) {
		t.Parallel()

		registeredProcessors := map[chainP2P.NetworkType]map[string]uint32{
			chainP2P.MainNetwork:         make(map[string]uint32),
			chainP2P.TransactionsNetwork: make(map[string]uint32),
		}
		args := createArgs()
		args.MainMessenger = &p2pmocks.MessengerStub{
			RegisterMessageProcessorCalled: func(networkType chainP2P.NetworkType, topic string, identifier string, handler chainP2P.MessageProcessor) error {
				registeredProcessors[networkType][topic]++

				return nil
			},
			HasTopicCalled: func(name string) bool {
				return true // extra coverage
			},
		}
		args.FullArchiveMessenger = &p2pmocks.MessengerStub{
			RegisterMessageProcessorCalled: func(networkType chainP2P.NetworkType, topic string, identifier string, handler chainP2P.MessageProcessor) error {
				registeredProcessors[networkType][topic]++

				return nil
			},
			HasTopicCalled: func(name string) bool {
				return true // extra coverage
			},
		}
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			NoShards: 2,
		}
		args.ResolversContainer = &dataRetriever.ResolversContainerStub{
			GetCalled: func(key string) (retriever.Resolver, error) {
				return &processP2PMocks.TopicResolverMock{
					RequestTopicCalled: func() string {
						return fmt.Sprintf("%s_REQUEST", key)
					},
				}, nil
			},
		}
		args.IsFullArchive = true // extra coverage

		handler, _ := p2p.NewEpochChangeTopicsHandler(args)
		require.NotNil(t, handler)

		handler.EpochConfirmed(1, 0)

		require.Equal(t, 2, len(registeredProcessors))
		mainProcessors := registeredProcessors[chainP2P.MainNetwork]
		require.Equal(t, 0, len(mainProcessors)) // should not register on main network

		transactionsProcessors := registeredProcessors[chainP2P.TransactionsNetwork]
		require.Equal(t, 12, len(transactionsProcessors))
		for _, cnt := range transactionsProcessors {
			// should have been called twice, full archive
			require.Equal(t, uint32(2), cnt)
		}
	})
	t.Run("should work and replace all topics, meta", func(t *testing.T) {
		t.Parallel()

		registeredProcessors := map[chainP2P.NetworkType]map[string]uint32{
			chainP2P.MainNetwork:         make(map[string]uint32),
			chainP2P.TransactionsNetwork: make(map[string]uint32),
		}
		args := createArgs()
		args.MainMessenger = &p2pmocks.MessengerStub{
			RegisterMessageProcessorCalled: func(networkType chainP2P.NetworkType, topic string, identifier string, handler chainP2P.MessageProcessor) error {
				registeredProcessors[networkType][topic]++

				return nil
			},
			HasTopicCalled: func(name string) bool {
				return true // extra coverage
			},
		}
		args.FullArchiveMessenger = &p2pmocks.MessengerStub{
			RegisterMessageProcessorCalled: func(networkType chainP2P.NetworkType, topic string, identifier string, handler chainP2P.MessageProcessor) error {
				registeredProcessors[networkType][topic]++

				return nil
			},
			HasTopicCalled: func(name string) bool {
				return true // extra coverage
			},
		}
		args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
			NoShards: 2,
			SelfIDCalled: func() uint32 {
				return core.MetachainShardId
			},
			CurrentShard: core.MetachainShardId,
		}
		args.ResolversContainer = &dataRetriever.ResolversContainerStub{
			GetCalled: func(key string) (retriever.Resolver, error) {
				return &processP2PMocks.TopicResolverMock{
					RequestTopicCalled: func() string {
						return fmt.Sprintf("%s_REQUEST", key)
					},
				}, nil
			},
		}

		handler, _ := p2p.NewEpochChangeTopicsHandler(args)
		require.NotNil(t, handler)

		handler.EpochConfirmed(1, 0)

		require.Equal(t, 2, len(registeredProcessors))
		mainProcessors := registeredProcessors[chainP2P.MainNetwork]
		require.Equal(t, 0, len(mainProcessors)) // should not register on main network

		transactionsProcessors := registeredProcessors[chainP2P.TransactionsNetwork]
		require.Equal(t, 14, len(transactionsProcessors))
		for _, cnt := range transactionsProcessors {
			// should have been called only once, no full archive
			require.Equal(t, uint32(1), cnt)
		}
	})
}
