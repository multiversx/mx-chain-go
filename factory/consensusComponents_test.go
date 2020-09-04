package factory_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/chronology"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/sposFactory"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	errorsErd "github.com/ElrondNetwork/elrond-go/errors"
	"github.com/ElrondNetwork/elrond-go/factory"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

// ------------ Test ConsensusComponentsFactory --------------------
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

//------------ Test Old Use Cases --------------------
func TestConsensusComponentsFactory_Create_GenesisBlockNotInitializedShouldErr(t *testing.T) {
	t.Parallel()

	consensusArgs := getConsensusArgs()
	consensusComponentsFactory, _ := factory.NewConsensusComponentsFactory(consensusArgs)
	managedConsensusComponents, _ := factory.NewManagedConsensusComponents(consensusComponentsFactory)

	dataComponents := consensusArgs.DataComponents

	dataComponents.SetBlockchain(&mock.ChainHandlerStub{
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

func TestConsensusComponentsFactory_Create_NilRounder(t *testing.T) {
	t.Parallel()

	consensusArgs := getConsensusArgs()
	processComponents := &mock.ProcessComponentsMock{}
	consensusArgs.ProcessComponents = processComponents
	consensusComponentsFactory, _ := factory.NewConsensusComponentsFactory(consensusArgs)

	cc, err := consensusComponentsFactory.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilShardCoordinator, err)
}

func TestConsensusComponentsFactory_Create_ConsensusTopicValidatorAlreadySet(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &mock.MessengerStub{
		HasTopicValidatorCalled: func(name string) bool {
			return true
		},
		HasTopicCalled: func(name string) bool {
			return true
		},
	}
	args.NetworkComponents = networkComponents

	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrValidatorAlreadySet, err)
}

func TestConsensusComponentsFactory_Create_ConsensusTopicCreateTopicError(t *testing.T) {
	t.Parallel()

	localError := errors.New("error")
	args := getConsensusArgs()
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = &mock.MessengerStub{
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

	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, localError, err)
}

func TestConsensusComponentsFactory_Create_ConsensusTopicNilMessageProcessor(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	networkComponents := getDefaultNetworkComponents()
	networkComponents.Messenger = nil
	args.NetworkComponents = networkComponents

	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilMessenger, err)
}

func TestConsensusComponentsFactory_Create_NilSyncTimer(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	coreComponents := getDefaultCoreComponents()
	coreComponents.NtpSyncTimer = nil
	args.CoreComponents = coreComponents
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, chronology.ErrNilSyncTimer, err)
}

func TestStartConsensus_ShardBootstrapperNilAccounts(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	stateComponents := getDefaultStateComponents()
	stateComponents.Accounts = nil
	args.StateComponents = stateComponents
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, process.ErrNilAccountsAdapter, err)
}

func TestStartConsensus_ShardBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = nil
	args.DataComponents = dataComponents
	processComponents := getDefaultProcessComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = 0
	processComponents.ShardCoord = shardCoordinator
	args.ProcessComponents = processComponents
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilDataPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperNilPoolHolder(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	dataComponents := getDefaultDataComponents()
	dataComponents.DataPool = nil
	args.DataComponents = dataComponents
	processComponents := getDefaultProcessComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = core.MetachainShardId
	processComponents.ShardCoord = shardCoordinator
	args.ProcessComponents = processComponents
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, errorsErd.ErrNilDataPoolsHolder, err)
}

func TestStartConsensus_MetaBootstrapperWrongNumberShards(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	processComponents := getDefaultProcessComponents()
	shardCoordinator := mock.NewMultiShardsCoordinatorMock(1)
	shardCoordinator.CurrentShard = 2
	processComponents.ShardCoord = shardCoordinator
	args.ProcessComponents = processComponents
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()

	require.Nil(t, cc)
	require.Equal(t, sharding.ErrShardIdOutOfRange, err)
}

func TestStartConsensus_ShardBootstrapperPubKeyToByteArrayError(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	args := getConsensusArgs()
	cryptoParams := getDefaultCryptoComponents()
	cryptoParams.PubKey = &mock.PublicKeyMock{
		ToByteArrayHandler: func() (i []byte, err error) {
			return []byte("nil"), localErr
		},
	}
	args.CryptoComponents = cryptoParams
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()
	require.Nil(t, cc)
	require.Equal(t, localErr, err)
}

func TestStartConsensus_ShardBootstrapperInvalidConsensusType(t *testing.T) {
	t.Parallel()

	args := getConsensusArgs()
	args.Config.Consensus.Type = "invalid"
	bcf, _ := factory.NewConsensusComponentsFactory(args)
	cc, err := bcf.Create()
	require.Nil(t, cc)
	require.Equal(t, sposFactory.ErrInvalidConsensusType, err)
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

func getDefaultNetworkComponents() *mock.NetworkComponentsMock {
	return &mock.NetworkComponentsMock{
		Messenger:       &mock.MessengerStub{},
		InputAntiFlood:  &mock.P2PAntifloodHandlerStub{},
		OutputAntiFlood: &mock.P2PAntifloodHandlerStub{},
		PeerBlackList:   &mock.PeerBlackListHandlerStub{},
	}
}

func getDefaultStateComponents() *testscommon.StateComponentsMock {
	return &testscommon.StateComponentsMock{
		PeersAcc:        &mock.AccountsStub{},
		Accounts:        &mock.AccountsStub{},
		Tries:           &mock.TriesHolderStub{},
		StorageManagers: map[string]data.StorageManager{"0": &mock.StorageManagerStub{}},
	}
}

func getDefaultDataComponents() *mock.DataComponentsMock {
	return &mock.DataComponentsMock{
		Blkc:              &mock.ChainHandlerStub{},
		Storage:           &mock.ChainStorerStub{},
		DataPool:          &testscommon.PoolsHolderMock{},
		MiniBlockProvider: &mock.MiniBlocksProviderStub{},
	}
}

func getDefaultProcessComponents() *mock.ProcessComponentsMock {
	return &mock.ProcessComponentsMock{
		NodesCoord: &mock.NodesCoordinatorMock{},
		ShardCoord: &testscommon.ShardsCoordinatorMock{
			NoShards:     1,
			CurrentShard: 0,
		},
		IntContainer:             &mock.InterceptorsContainerStub{},
		ResFinder:                &mock.ResolversFinderStub{},
		RoundHandler:             &testscommon.RounderMock{},
		EpochTrigger:             &testscommon.EpochStartTriggerStub{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		ForkDetect:               &mock.ForkDetectorMock{},
		BlockProcess:             &mock.BlockProcessorStub{},
		BlackListHdl:             &testscommon.TimeCacheStub{},
		BootSore:                 &mock.BootstrapStorerMock{},
		HeaderSigVerif:           &mock.HeaderSigVerifierStub{},
		HeaderIntegrVerif:        &mock.HeaderIntegrityVerifierStub{},
		ValidatorStatistics:      &mock.ValidatorStatisticsProcessorStub{},
		ValidatorProvider:        &mock.ValidatorsProviderStub{},
		BlockTrack:               &mock.BlockTrackerStub{},
		PendingMiniBlocksHdl:     &mock.PendingMiniBlocksHandlerStub{},
		ReqHandler:               &mock.RequestHandlerStub{},
		TxLogsProcess:            &mock.TxLogProcessorMock{},
		HeaderConstructValidator: &mock.HeaderValidatorStub{},
		PeerMapper:               &mock.NetworkShardingCollectorStub{},
	}
}

func getDefaultCryptoComponents() *mock.CryptoComponentsMock {
	return &mock.CryptoComponentsMock{
		PubKey:          &mock.PublicKeyMock{},
		PrivKey:         &mock.PrivateKeyStub{},
		PubKeyString:    "pubKey",
		PrivKeyBytes:    []byte("privKey"),
		PubKeyBytes:     []byte("pubKey"),
		BlockSig:        &mock.SinglesignMock{},
		TxSig:           &mock.SinglesignMock{},
		MultiSig:        &mock.MultisignMock{},
		PeerSignHandler: &mock.PeerSignatureHandler{},
		BlKeyGen:        &mock.KeyGenMock{},
		TxKeyGen:        &mock.KeyGenMock{},
		MsgSigVerifier:  &testscommon.MessageSignVerifierMock{},
	}
}
