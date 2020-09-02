package factory_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

// ------------ Test BootstrapComponentsFactory --------------------
func TestNewConsensusComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.NotNil(t, bcf)
	require.Nil(t, err)
}

func TestNewConsensusComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.CoreComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCoreComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilDataComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.DataComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilDataComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilCryptoComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.CryptoComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCryptoComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilNetworkComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.NetworkComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilNetworkComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilProcessComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.ProcessComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilProcessComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilStateComponents(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.StateComponents = nil

	bcf, err := factory.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilStateComponentsHolder, err)
}

func TestConsensusComponentsFactory_CreateForShard(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	ccf, _ := factory.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

type wrappedProcessComponents struct {
	factory.ProcessComponentsHolder
}

func (wp *wrappedProcessComponents) ShardCoordinator() sharding.Coordinator {
	shC := mock.NewMultiShardsCoordinatorMock(2)
	shC.SelfIDCalled = func() uint32 {
		return core.MetachainShardId
	}

	return shC
}

func TestConsensusComponentsFactory_CreateForMeta(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()

	args.ProcessComponents = &wrappedProcessComponents{
		ProcessComponentsHolder: args.ProcessComponents,
	}
	ccf, _ := factory.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func getConsensusArgs() factory.ConsensusComponentsFactoryArgs {
	coreComponents := getCoreComponents()
	networkComponents := getNetworkComponents()
	stateComponents := getStateComponents(coreComponents)
	cryptoComponents := getCryptoComponents(coreComponents)
	dataComponents := getDataComponents(coreComponents)
	processComponents := getProcessComponents(
		coreComponents,
		networkComponents,
		dataComponents,
		cryptoComponents,
		stateComponents,
	)
	statusComponents := getStatusComponents(
		coreComponents,
		networkComponents,
		dataComponents,
		processComponents,
	)

	return factory.ConsensusComponentsFactoryArgs{
		Config:              testscommon.GetGeneralConfig(),
		ConsensusGroupSize:  5,
		BootstrapRoundIndex: 0,
		HardforkTrigger:     &mock.HardforkTriggerStub{},
		CoreComponents:      coreComponents,
		NetworkComponents:   networkComponents,
		CryptoComponents:    cryptoComponents,
		DataComponents:      dataComponents,
		ProcessComponents:   processComponents,
		StateComponents:     stateComponents,
		StatusComponents:    statusComponents,
	}
}
