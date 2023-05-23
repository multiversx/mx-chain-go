package heartbeat_test

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	heartbeatComp "github.com/multiversx/mx-chain-go/factory/heartbeat"
	testsMocks "github.com/multiversx/mx-chain-go/integrationTests/mock"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapMocks"
	componentsMock "github.com/multiversx/mx-chain-go/testscommon/components"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/mainFactoryMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

func createMockHeartbeatV2ComponentsFactoryArgs() heartbeatComp.ArgHeartbeatV2ComponentsFactory {
	return heartbeatComp.ArgHeartbeatV2ComponentsFactory{
		Config: createMockConfig(),
		Prefs: config.Preferences{
			Preferences: config.PreferencesConfig{
				NodeDisplayName: "node",
				Identity:        "identity",
			},
		},
		AppVersion: "test",
		BootstrapComponents: &mainFactoryMocks.BootstrapComponentsStub{
			ShCoordinator:   &testscommon.ShardsCoordinatorMock{},
			BootstrapParams: &bootstrapMocks.BootstrapParamsHandlerMock{},
		},
		CoreComponents: &factory.CoreComponentsHolderStub{
			InternalMarshalizerCalled: func() marshal.Marshalizer {
				return &marshallerMock.MarshalizerStub{}
			},
			HardforkTriggerPubKeyCalled: func() []byte {
				return []byte("hardfork pub key")
			},
			ValidatorPubKeyConverterCalled: func() core.PubkeyConverter {
				return &testscommon.PubkeyConverterStub{}
			},
		},
		DataComponents: &testsMocks.DataComponentsStub{
			DataPool: &dataRetriever.PoolsHolderStub{
				PeerAuthenticationsCalled: func() storage.Cacher {
					return &testscommon.CacherStub{}
				},
				HeartbeatsCalled: func() storage.Cacher {
					return &testscommon.CacherStub{}
				},
			},
			BlockChain: &testscommon.ChainHandlerStub{},
		},
		NetworkComponents: &testsMocks.NetworkComponentsStub{
			Messenger: &p2pmocks.MessengerStub{},
		},
		CryptoComponents: &testsMocks.CryptoComponentsStub{
			PrivKey:                 &cryptoMocks.PrivateKeyStub{},
			PeerSignHandler:         &testsMocks.PeerSignatureHandler{},
			ManagedPeersHolderField: &testscommon.ManagedPeersHolderStub{},
		},
		ProcessComponents: &testsMocks.ProcessComponentsStub{
			EpochTrigger:                  &testsMocks.EpochStartTriggerStub{},
			EpochNotifier:                 &testsMocks.EpochStartNotifierStub{},
			NodesCoord:                    &shardingMocks.NodesCoordinatorStub{},
			NodeRedundancyHandlerInternal: &testsMocks.RedundancyHandlerStub{},
			HardforkTriggerField:          &testscommon.HardforkTriggerStub{},
			ReqHandler:                    &testscommon.RequestHandlerStub{},
			PeerMapper:                    &testsMocks.PeerShardMapperStub{},
			ShardCoord:                    &testscommon.ShardsCoordinatorMock{},
		},
		StatusCoreComponents: &factory.StatusCoreComponentsStub{
			AppStatusHandlerField: &statusHandler.AppStatusHandlerStub{},
		},
	}
}

func createMockConfig() config.Config {
	return config.Config{
		HeartbeatV2: config.HeartbeatV2Config{
			PeerAuthenticationTimeBetweenSendsInSec:          1,
			PeerAuthenticationTimeBetweenSendsWhenErrorInSec: 1,
			PeerAuthenticationTimeThresholdBetweenSends:      0.1,
			HeartbeatTimeBetweenSendsInSec:                   1,
			HeartbeatTimeBetweenSendsDuringBootstrapInSec:    1,
			HeartbeatTimeBetweenSendsWhenErrorInSec:          1,
			HeartbeatTimeThresholdBetweenSends:               0.1,
			HeartbeatExpiryTimespanInSec:                     30,
			MinPeersThreshold:                                0.8,
			DelayBetweenPeerAuthenticationRequestsInSec:      10,
			PeerAuthenticationMaxTimeoutForRequestsInSec:     60,
			PeerAuthenticationTimeBetweenChecksInSec:         1,
			PeerShardTimeBetweenSendsInSec:                   5,
			PeerShardTimeThresholdBetweenSends:               0.1,
			MaxMissingKeysInRequest:                          100,
			MaxDurationPeerUnresponsiveInSec:                 10,
			HideInactiveValidatorIntervalInSec:               60,
			HardforkTimeBetweenSendsInSec:                    5,
			TimeBetweenConnectionsMetricsUpdateInSec:         10,
			TimeToReadDirectConnectionsInSec:                 15,
			HeartbeatPool: config.CacheConfig{
				Type:     "LRU",
				Capacity: 1000,
				Shards:   1,
			},
		},
		Hardfork: config.HardforkConfig{
			PublicKeyToListenFrom: componentsMock.DummyPk,
		},
	}
}

func TestNewHeartbeatV2ComponentsFactory(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(createMockHeartbeatV2ComponentsFactoryArgs())
		assert.NotNil(t, hcf)
		assert.NoError(t, err)
	})
	t.Run("nil BootstrapComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.BootstrapComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilBootstrapComponentsHolder, err)
	})
	t.Run("nil CoreComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.CoreComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilCoreComponentsHolder, err)
	})
	t.Run("nil DataComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.DataComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilDataComponentsHolder, err)
	})
	t.Run("nil DataPool should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.DataComponents = &testsMocks.DataComponentsStub{
			DataPool: nil,
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilDataPoolsHolder, err)
	})
	t.Run("nil NetworkComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilNetworkComponentsHolder, err)
	})
	t.Run("nil NetworkMessenger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = &testsMocks.NetworkComponentsStub{
			Messenger: nil,
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilMessenger, err)
	})
	t.Run("nil CryptoComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.CryptoComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilCryptoComponentsHolder, err)
	})
	t.Run("nil ProcessComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.ProcessComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilProcessComponentsHolder, err)
	})
	t.Run("nil EpochStartTrigger should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.ProcessComponents = &testsMocks.ProcessComponentsStub{
			EpochTrigger: nil,
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilEpochStartTrigger, err)
	})
	t.Run("nil StatusCoreComponents should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.StatusCoreComponents = nil
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.Nil(t, hcf)
		assert.Equal(t, errorsMx.ErrNilStatusCoreComponents, err)
	})
}

func TestHeartbeatV2Components_Create(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	t.Run("messenger does not have PeerAuthenticationTopic and fails to create it", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = &testsMocks.NetworkComponentsStub{
			Messenger: &p2pmocks.MessengerStub{
				HasTopicCalled: func(name string) bool {
					if name == common.PeerAuthenticationTopic {
						return false
					}
					assert.Fail(t, "should not have been called")
					return true
				},
				CreateTopicCalled: func(name string, createChannelForTopic bool) error {
					if name == common.PeerAuthenticationTopic {
						return expectedErr
					}
					assert.Fail(t, "should not have been called")
					return nil
				},
			},
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("messenger does not have HeartbeatV2Topic and fails to create it", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = &testsMocks.NetworkComponentsStub{
			Messenger: &p2pmocks.MessengerStub{
				HasTopicCalled: func(name string) bool {
					return name != common.HeartbeatV2Topic
				},
				CreateTopicCalled: func(name string, createChannelForTopic bool) error {
					if name == common.HeartbeatV2Topic {
						return expectedErr
					}
					assert.Fail(t, "should not have been called")
					return nil
				},
			},
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("invalid config should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.HeartbeatExpiryTimespanInSec = args.Config.HeartbeatV2.PeerAuthenticationTimeBetweenSendsInSec
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.True(t, errors.Is(err, errorsMx.ErrInvalidHeartbeatV2Config))
	})
	t.Run("NewPeerTypeProvider fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		processComp := args.ProcessComponents
		args.ProcessComponents = &testsMocks.ProcessComponentsStub{
			NodesCoord:    nil,
			EpochTrigger:  processComp.EpochStartTrigger(),
			EpochNotifier: processComp.EpochStartNotifier(),
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewSender fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.PeerAuthenticationTimeBetweenSendsInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewPeerAuthenticationRequestsProcessor fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.DelayBetweenPeerAuthenticationRequestsInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewPeerShardSender fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.PeerShardTimeBetweenSendsInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewHeartbeatV2Monitor fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.MaxDurationPeerUnresponsiveInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewMetricsUpdater fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.TimeBetweenConnectionsMetricsUpdateInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewDirectConnectionProcessor fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Config.HeartbeatV2.TimeToReadDirectConnectionsInSec = 0
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("NewCrossShardPeerTopicNotifier fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		processComp := args.ProcessComponents
		cnt := 0
		args.ProcessComponents = &testsMocks.ProcessComponentsStub{
			NodesCoord:                    processComp.NodesCoordinator(),
			EpochTrigger:                  processComp.EpochStartTrigger(),
			EpochNotifier:                 processComp.EpochStartNotifier(),
			NodeRedundancyHandlerInternal: processComp.NodeRedundancyHandler(),
			HardforkTriggerField:          processComp.HardforkTrigger(),
			PeerMapper:                    processComp.PeerShardMapper(),
			ShardCoordinatorCalled: func() sharding.Coordinator {
				cnt++
				if cnt > 3 {
					return nil
				}
				return processComp.ShardCoordinator()
			},
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Error(t, err)
	})
	t.Run("AddPeerTopicNotifier fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.NetworkComponents = &testsMocks.NetworkComponentsStub{
			Messenger: &p2pmocks.MessengerStub{
				AddPeerTopicNotifierCalled: func(notifier p2p.PeerTopicNotifier) error {
					return expectedErr
				},
			},
		}
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.Nil(t, hc)
		assert.Equal(t, expectedErr, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		defer func() {
			r := recover()
			if r != nil {
				assert.Fail(t, "should not panic")
			}
		}()

		args := createMockHeartbeatV2ComponentsFactoryArgs()
		args.Prefs.Preferences.FullArchive = true // coverage only
		hcf, err := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
		assert.NotNil(t, hcf)
		assert.NoError(t, err)

		hc, err := hcf.Create()
		assert.NotNil(t, hc)
		assert.NoError(t, err)
		assert.NoError(t, hc.Close())
	})
}

func TestHeartbeatV2ComponentsFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	args := createMockHeartbeatV2ComponentsFactoryArgs()
	args.CoreComponents = nil
	hcf, _ := heartbeatComp.NewHeartbeatV2ComponentsFactory(args)
	assert.True(t, hcf.IsInterfaceNil())

	hcf, _ = heartbeatComp.NewHeartbeatV2ComponentsFactory(createMockHeartbeatV2ComponentsFactoryArgs())
	assert.False(t, hcf.IsInterfaceNil())
}
