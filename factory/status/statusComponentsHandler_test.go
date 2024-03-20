package status_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory"
	"github.com/multiversx/mx-chain-go/factory/mock"
	statusComp "github.com/multiversx/mx-chain-go/factory/status"
	"github.com/multiversx/mx-chain-go/p2p"
	factoryMocks "github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/require"
)

func TestNewManagedStatusComponents(t *testing.T) {
	t.Parallel()

	t.Run("nil factory should error", func(t *testing.T) {
		t.Parallel()

		managedStatusComponents, err := statusComp.NewManagedStatusComponents(nil)
		require.Equal(t, errorsMx.ErrNilStatusComponentsFactory, err)
		require.Nil(t, managedStatusComponents)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scf, err := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
		require.Nil(t, err)
		managedStatusComponents, err := statusComp.NewManagedStatusComponents(scf)
		require.Nil(t, err)
		require.NotNil(t, managedStatusComponents)
	})
}

func TestManagedStatusComponents_Create(t *testing.T) {
	t.Parallel()

	t.Run("invalid params should error", func(t *testing.T) {
		t.Parallel()

		args := createMockStatusComponentsFactoryArgs()
		args.StatusCoreComponents = &factoryMocks.StatusCoreComponentsStub{
			AppStatusHandlerField: nil,
		}
		scf, err := statusComp.NewStatusComponentsFactory(args)
		require.Nil(t, err)
		managedStatusComponents, err := statusComp.NewManagedStatusComponents(scf)
		require.Nil(t, err)
		require.NotNil(t, managedStatusComponents)

		err = managedStatusComponents.Create()
		require.Error(t, err)
	})
	t.Run("should work with getters", func(t *testing.T) {
		t.Parallel()

		scf, err := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
		require.Nil(t, err)
		managedStatusComponents, err := statusComp.NewManagedStatusComponents(scf)
		require.Nil(t, err)
		require.NotNil(t, managedStatusComponents)
		require.Nil(t, managedStatusComponents.OutportHandler())
		require.Nil(t, managedStatusComponents.SoftwareVersionChecker())
		require.Nil(t, managedStatusComponents.ManagedPeersMonitor())

		err = managedStatusComponents.Create()
		require.NoError(t, err)
		require.NotNil(t, managedStatusComponents.OutportHandler())
		require.NotNil(t, managedStatusComponents.SoftwareVersionChecker())
		require.NotNil(t, managedStatusComponents.ManagedPeersMonitor())

		require.Equal(t, factory.StatusComponentsName, managedStatusComponents.String())
	})
}

func TestManagedStatusComponents_Close(t *testing.T) {
	t.Parallel()

	scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)
	err := managedStatusComponents.Close()
	require.NoError(t, err)

	err = managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.StartPolling() // coverage
	require.NoError(t, err)

	err = managedStatusComponents.Close()
	require.NoError(t, err)
}

func TestManagedStatusComponents_CheckSubcomponents(t *testing.T) {
	t.Parallel()

	scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)

	err := managedStatusComponents.CheckSubcomponents()
	require.Equal(t, errorsMx.ErrNilStatusComponents, err)

	err = managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.CheckSubcomponents()
	require.NoError(t, err)
}

func TestManagedStatusComponents_SetForkDetector(t *testing.T) {
	t.Parallel()

	scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)
	err := managedStatusComponents.Create()
	require.NoError(t, err)

	err = managedStatusComponents.SetForkDetector(nil)
	require.Equal(t, errorsMx.ErrNilForkDetector, err)
	err = managedStatusComponents.SetForkDetector(&mock.ForkDetectorMock{})
	require.NoError(t, err)
}

func TestManagedStatusComponents_StartPolling(t *testing.T) {
	t.Parallel()

	t.Run("NewAppStatusPolling fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockStatusComponentsFactoryArgs()
		args.Config.GeneralSettings.StatusPollingIntervalSec = 0
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)
		err := managedStatusComponents.Create()
		require.NoError(t, err)

		err = managedStatusComponents.StartPolling()
		require.Equal(t, errorsMx.ErrStatusPollingInit, err)
	})
	t.Run("RegisterPollingFunc fails should error", func(t *testing.T) {
		t.Parallel()

		args := createMockStatusComponentsFactoryArgs()
		args.Config.GeneralSettings.StatusPollingIntervalSec = 0
		scf, _ := statusComp.NewStatusComponentsFactory(args)
		managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)
		err := managedStatusComponents.Create()
		require.NoError(t, err)

		err = managedStatusComponents.StartPolling()
		require.Equal(t, errorsMx.ErrStatusPollingInit, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
		managedStatusComponents, _ := statusComp.NewManagedStatusComponents(scf)
		err := managedStatusComponents.Create()
		require.NoError(t, err)
		require.NoError(t, managedStatusComponents.SetForkDetector(&mock.ForkDetectorStub{}))

		err = managedStatusComponents.StartPolling()
		require.NoError(t, err)
	})
}

func TestComputeNumConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("main network", testComputeNumConnectedPeers(""))
	t.Run("full archive network", testComputeNumConnectedPeers(common.FullArchiveMetricSuffix))
}

func testComputeNumConnectedPeers(suffix string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		netMes := &p2pmocks.MessengerStub{
			ConnectedAddressesCalled: func() []string {
				return []string{"addr1", "addr2", "addr3"}
			},
		}
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetUInt64ValueHandler: func(key string, value uint64) {
				require.Equal(t, common.SuffixedMetric(common.MetricNumConnectedPeers, suffix), key)
				require.Equal(t, uint64(3), value)
			},
		}

		statusComp.ComputeNumConnectedPeers(appStatusHandler, netMes, suffix)
	}
}

func TestComputeConnectedPeers(t *testing.T) {
	t.Parallel()

	t.Run("main network", testComputeConnectedPeers(""))
	t.Run("full archive network", testComputeConnectedPeers(common.FullArchiveMetricSuffix))
}

func testComputeConnectedPeers(suffix string) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		netMes := &p2pmocks.MessengerStub{
			GetConnectedPeersInfoCalled: func() *p2p.ConnectedPeersInfo {
				return &p2p.ConnectedPeersInfo{
					SelfShardID:  0,
					UnknownPeers: []string{"unknown"},
					Seeders:      []string{"seeder"},
					IntraShardValidators: map[uint32][]string{
						0: {"intra-v-0"},
						1: {"intra-v-1"},
					},
					IntraShardObservers: map[uint32][]string{
						0: {"intra-o-0"},
						1: {"intra-o-1"},
					},
					CrossShardValidators: map[uint32][]string{
						0: {"cross-v-0"},
						1: {"cross-v-1"},
					},
					CrossShardObservers: map[uint32][]string{
						0: {"cross-o-0"},
						1: {"cross-o-1"},
					},
					NumValidatorsOnShard: map[uint32]int{
						0: 1,
						1: 1,
					},
					NumObserversOnShard: map[uint32]int{
						0: 1,
						1: 1,
					},
					NumPreferredPeersOnShard: map[uint32]int{
						0: 0,
						1: 0,
					},
					NumIntraShardValidators: 2,
					NumIntraShardObservers:  2,
					NumCrossShardValidators: 2,
					NumCrossShardObservers:  2,
				}
			},
			AddressesCalled: func() []string {
				return []string{"intra-v-0", "intra-v-1", "intra-o-0", "intra-o-1", "cross-v-0", "cross-v-1"}
			},
		}
		expectedPeerClassification := "intraVal:2,crossVal:2,intraObs:2,crossObs:2,unknown:1,"
		cnt := 0
		appStatusHandler := &statusHandler.AppStatusHandlerStub{
			SetStringValueHandler: func(key string, value string) {
				cnt++
				switch cnt {
				case 1:
					require.Equal(t, common.SuffixedMetric(common.MetricNumConnectedPeersClassification, suffix), key)
					require.Equal(t, expectedPeerClassification, value)
				case 2:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PNumConnectedPeersClassification, suffix), key)
					require.Equal(t, expectedPeerClassification, value)
				case 3:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PUnknownPeers, suffix), key)
					require.Equal(t, "unknown", value)
				case 4:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PIntraShardValidators, suffix), key)
					require.Equal(t, "intra-v-0,intra-v-1", value)
				case 5:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PIntraShardObservers, suffix), key)
					require.Equal(t, "intra-o-0,intra-o-1", value)
				case 6:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PCrossShardValidators, suffix), key)
					require.Equal(t, "cross-v-0,cross-v-1", value)
				case 7:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PCrossShardObservers, suffix), key)
					require.Equal(t, "cross-o-0,cross-o-1", value)
				case 8:
					require.Equal(t, common.SuffixedMetric(common.MetricP2PPeerInfo, suffix), key)
					require.Equal(t, "intra-v-0,intra-v-1,intra-o-0,intra-o-1,cross-v-0,cross-v-1", value)
				default:
					require.Fail(t, "should not have been called")
				}
			},
			SetUInt64ValueHandler: func(key string, value uint64) {
				require.Equal(t, common.SuffixedMetric(common.MetricNumConnectedPeers, suffix), key)
				require.Equal(t, 3, key)
			},
		}

		statusComp.ComputeConnectedPeers(appStatusHandler, netMes, suffix)
	}
}

func TestManagedStatusComponents_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	managedStatusComponents, _ := statusComp.NewManagedStatusComponents(nil)
	require.True(t, managedStatusComponents.IsInterfaceNil())

	scf, _ := statusComp.NewStatusComponentsFactory(createMockStatusComponentsFactoryArgs())
	managedStatusComponents, _ = statusComp.NewManagedStatusComponents(scf)
	require.False(t, managedStatusComponents.IsInterfaceNil())
}
