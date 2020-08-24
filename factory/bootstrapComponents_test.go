package factory_test

import (
	"errors"
	"testing"
	"time"

	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

// ------------ Test BootstrapComponentsFactory --------------------
func TestNewBootstrapComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.NotNil(t, bcf)
	require.Nil(t, err)
}

func TestNewBootstrapComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.CoreComponents = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCoreComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilCryptoComponents(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.CryptoComponents = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCryptoComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilNetworkComponents(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.NetworkComponents = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilNetworkComponentsHolder, err)
}

func TestNewBootstrapComponentsFactory_NilNodeShuffler(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.NodeShuffler = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilShuffler, err)
}

func TestNewBootstrapComponentsFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.ShardCoordinator = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilShardCoordinator, err)
}

func TestNewBootstrapComponentsFactory_NilGenesisNodesSetup(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.GenesisNodesSetup = nil

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilGenesisNodesSetup, err)
}

func TestNewBootstrapComponentsFactory_NilWorkingDir(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	args.WorkingDir = ""

	bcf, err := factory.NewBootstrapComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrInvalidWorkingDir, err)
}

func TestBootstrapComponentsFactory_Create_ShouldWork(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()

	bcf, _ := factory.NewBootstrapComponentsFactory(args)

	bc, err := bcf.Create()

	require.NotNil(t, bc)
	require.Nil(t, err)
}

func TestBootstrapComponentsFactory_Create_BootstrapDataProviderCreationFail(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	coreComponents := getDefaultCoreComponents()
	args.CoreComponents = coreComponents

	bcf, _ := factory.NewBootstrapComponentsFactory(args)

	coreComponents.IntMarsh = nil
	bc, err := bcf.Create()

	require.Nil(t, bc)
	require.True(t, errors.Is(err, errorsErd.ErrBootstrapDataProviderCreationFailed))
}

func TestBootstrapComponentsFactory_Create_NilHasher(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()

	bcf, _ := factory.NewBootstrapComponentsFactory(args)

	bc, err := bcf.Create()

	require.Nil(t, bc)
	require.Equal(t, errorsErd.ErrNilMarshalizer, err)
}

func getBootStrapArgs() factory.BootstrapComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	cryptoComponents := getCryptoComponents(coreComponents)
	return factory.BootstrapComponentsFactoryArgs{
		Config:                testscommon.GetGeneralConfig(),
		WorkingDir:            "home",
		DestinationAsObserver: 0,
		GenesisNodesSetup:     &mock.NodesSetupStub{},
		NodeShuffler:          &mock.NodeShufflerMock{},
		ShardCoordinator:      mock.NewMultiShardsCoordinatorMock(2),
		CoreComponents:        coreComponents,
		CryptoComponents:      cryptoComponents,
		NetworkComponents:     networkComponents,
	}
}

func getDefaultCoreComponents() *mock.CoreComponentsMock {
	return &mock.CoreComponentsMock{
		IntMarsh:            &testscommon.MarshalizerMock{},
		TxMarsh:             &testscommon.MarshalizerMock{},
		VmMarsh:             &testscommon.MarshalizerMock{},
		Hash:                &testscommon.HasherMock{},
		UInt64ByteSliceConv: testscommon.NewNonceHashConverterMock(),
		AddrPubKeyConv:      testscommon.NewPubkeyConverterMock(32),
		ValPubKeyConv:       testscommon.NewPubkeyConverterMock(32),
		PathHdl:             &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return "chainID"
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		StatusHdl:        &testscommon.AppStatusHandlerStub{},
		WatchdogTimer:    &testscommon.WatchdogMock{},
		AlarmSch:         &testscommon.AlarmSchedulerStub{},
		NtpSyncTimer:     &testscommon.SyncTimerStub{},
		RoundHandler:     &testscommon.RounderMock{},
		EconomicsHandler: &testscommon.EconomicsHandlerMock{},
		RatingsConfig:    &testscommon.RatingsInfoMock{},
		RatingHandler:    &testscommon.RaterMock{},
		NodesConfig:      &testscommon.NodesSetupStub{},
		StartTime:        time.Time{},
	}
}
