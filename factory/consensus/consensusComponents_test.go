package consensus_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	consensusComp "github.com/ElrondNetwork/elrond-go/factory/consensus"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	componentsMock "github.com/ElrondNetwork/elrond-go/factory/mock/components"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/require"
)

// ------------ Test ConsensusComponentsFactory --------------------
func TestNewConsensusComponentsFactory_OkValuesShouldWork(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.NotNil(t, bcf)
	require.Nil(t, err)
}

func TestNewConsensusComponentsFactory_NilCoreComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.CoreComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCoreComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilDataComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.DataComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilDataComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilCryptoComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.CryptoComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilCryptoComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilNetworkComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.NetworkComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilNetworkComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilProcessComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.ProcessComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilProcessComponentsHolder, err)
}

func TestNewConsensusComponentsFactory_NilStateComponents(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.StateComponents = nil

	bcf, err := consensusComp.NewConsensusComponentsFactory(args)

	require.Nil(t, bcf)
	require.Equal(t, errorsErd.ErrNilStateComponentsHolder, err)
}

// ------------ Test Old Use Cases --------------------
func TestConsensusComponentsFactory_CreateGenesisBlockNotInitializedShouldErr(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	consensusArgs := componentsMock.GetConsensusArgs(shardCoordinator)
	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(consensusArgs)
	managedConsensusComponents, _ := consensusComp.NewManagedConsensusComponents(consensusComponentsFactory)

	dataComponents := consensusArgs.DataComponents

	dataComponents.SetBlockchain(&testscommon.ChainHandlerStub{
		GetGenesisHeaderHashCalled: func() []byte {
			return nil
		},
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return nil
		},
	})

	err := managedConsensusComponents.Create()
	require.True(t, errors.Is(err, errorsErd.ErrConsensusComponentsFactoryCreate))
	require.True(t, strings.Contains(err.Error(), errorsErd.ErrGenesisBlockNotInitialized.Error()))
}

func TestConsensusComponentsFactory_CreateForShard(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	ccf, _ := consensusComp.NewConsensusComponentsFactory(args)
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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)

	args.ProcessComponents = &wrappedProcessComponents{
		ProcessComponentsHolder: args.ProcessComponents,
	}
	ccf, _ := consensusComp.NewConsensusComponentsFactory(args)
	require.NotNil(t, ccf)

	cc, err := ccf.Create()
	require.NoError(t, err)
	require.NotNil(t, cc)
}

func TestConsensusComponentsFactory_CreateNilShardCoordinator(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	consensusArgs := componentsMock.GetConsensusArgs(shardCoordinator)
	processComponents := &mock.ProcessComponentsMock{}
	consensusArgs.ProcessComponents = processComponents
	consensusComponentsFactory, _ := consensusComp.NewConsensusComponentsFactory(consensusArgs)

	cc, err := consensusComponentsFactory.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilShardCoordinator, err)
}

func TestConsensusComponentsFactory_CreateConsensusTopicCreateTopicError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	localError := errors.New("error")
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	networkComponents := componentsMock.GetDefaultNetworkComponents()
	networkComponents.Messenger = &p2pmocks.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return false
		},
		HasTopicCalled: func(name string) bool {
			return false
		},
		CreateTopicCalled: func(name string, createChannelForTopic bool) error {
			return localError
		},
	}
	args.NetworkComponents = networkComponents

	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, localError, err)
}

func TestConsensusComponentsFactory_CreateConsensusTopicNilMessageProcessor(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	networkComponents := componentsMock.GetDefaultNetworkComponents()
	networkComponents.Messenger = nil
	args.NetworkComponents = networkComponents

	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilMessenger, err)
}

func TestConsensusComponentsFactory_CreateNilSyncTimer(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	coreComponents := componentsMock.GetDefaultCoreComponents()
	coreComponents.NtpSyncTimer = nil
	args.CoreComponents = coreComponents
	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, chronology.ErrNilSyncTimer, err)
}

func TestStartConsensus_ShardBootstrapperNilAccounts(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	stateComponents := componentsMock.GetDefaultStateComponents()
	stateComponents.Accounts = nil
	args.StateComponents = stateComponents
	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestStartConsensus_ShardBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = 0
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	dataComponents := componentsMock.GetDefaultDataComponents()
	dataComponents.DataPool = nil
	args.DataComponents = dataComponents
	processComponents := componentsMock.GetDefaultProcessComponents(shardCoordinator)
	args.ProcessComponents = processComponents
	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilDataPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = core.MetachainShardId
	shardCoordinator.ComputeIdCalled = func(address []byte) uint32 {
		if core.IsSmartContractOnMetachain(address[len(address)-1:], address) {
			return core.MetachainShardId
		}

		return 0
	}
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	dataComponents := componentsMock.GetDefaultDataComponents()
	dataComponents.DataPool = nil
	args.DataComponents = dataComponents
	args.ProcessComponents = componentsMock.GetDefaultProcessComponents(shardCoordinator)
	bcf, err := consensusComp.NewConsensusComponentsFactory(args)
	require.Nil(t, err)
	require.NotNil(t, bcf)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilDataPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperWrongNumberShards(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	processComponents := componentsMock.GetDefaultProcessComponents(shardCoordinator)
	args.ProcessComponents = processComponents
	bcf, err := consensusComp.NewConsensusComponentsFactory(args)
	require.Nil(t, err)
	shardCoordinator.CurrentShard = 2
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, sharding.ErrShardIdOutOfRange, err)
}

func TestStartConsensus_ShardBootstrapperPubKeyToByteArrayError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	localErr := errors.New("err")
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	cryptoParams := componentsMock.GetDefaultCryptoComponents()
	cryptoParams.PubKey = &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, err error) {
			return []byte("nil"), localErr
		},
	}
	args.CryptoComponents = cryptoParams
	bcf, _ := consensusComp.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()
	require.Nil(t, cc)
	require.Equal(t, localErr, err)
}

func TestStartConsensus_ShardBootstrapperInvalidConsensusType(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := mock.NewMultiShardsCoordinatorMock(2)
	args := componentsMock.GetConsensusArgs(shardCoordinator)
	args.Config.Consensus.Type = "invalid"
	bcf, err := consensusComp.NewConsensusComponentsFactory(args)
	require.Nil(t, err)
	cc, err := bcf.Create()
	require.Nil(t, cc)
	require.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}
