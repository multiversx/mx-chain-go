package factory_test

import (
	"errors"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/config"
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

	require.Nil(t, err)
	require.NotNil(t, bc)
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
	require.True(t, errors.Is(err, errorsErd.ErrNewBootstrapDataProvider))
}

func TestBootstrapComponentsFactory_Create_EpochStartBootstrapCreationFail(t *testing.T) {
	t.Parallel()

	args := getBootStrapArgs()
	coreComponents := getDefaultCoreComponents()
	args.CoreComponents = coreComponents

	bcf, _ := factory.NewBootstrapComponentsFactory(args)

	coreComponents.RatingHandler = nil
	bc, err := bcf.Create()

	require.Nil(t, bc)
	require.True(t, errors.Is(err, errorsErd.ErrNewEpochStartBootstrap))
}

func getBootStrapArgs() factory.BootstrapComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	cryptoComponents := getCryptoComponents(coreComponents)
	return factory.BootstrapComponentsFactoryArgs{
		Config:            testscommon.GetGeneralConfig(),
		WorkingDir:        "home",
		CoreComponents:    coreComponents,
		CryptoComponents:  cryptoComponents,
		NetworkComponents: networkComponents,
		PrefConfig: config.Preferences{
			Preferences: config.PreferencesConfig{
				DestinationShardAsObserver: "0",
			},
		},
		ImportDbConfig: config.ImportDbConfig{
			IsImportDBMode: false,
		},
	}
}

func getDefaultCoreComponents() *mock.CoreComponentsMock {
	return &mock.CoreComponentsMock{
		IntMarsh:            &testscommon.MarshalizerMock{},
		TxMarsh:             &testscommon.MarshalizerMock{},
		VmMarsh:             &testscommon.MarshalizerMock{},
		Hash:                &testscommon.HasherStub{},
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
		AppStatusHdl:  &testscommon.AppStatusHandlerStub{},
		WatchdogTimer: &testscommon.WatchdogMock{},
		AlarmSch:      &testscommon.AlarmSchedulerStub{},
		NtpSyncTimer:  &testscommon.SyncTimerStub{},
		RoundHandler:  &testscommon.RounderMock{},
		//TODO: uncomment this
		//EconomicsHandler: &testscommon.EconomicsHandlerMock{},
		RatingsConfig: &testscommon.RatingsInfoMock{},
		RatingHandler: &testscommon.RaterMock{},
		NodesConfig:   &testscommon.NodesSetupStub{},
		StartTime:     time.Time{},
	}
}
