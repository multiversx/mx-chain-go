package spos_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/bootstrapperStubs"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	epochstartmock "github.com/multiversx/mx-chain-go/testscommon/epochstartmock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

func initConsensusDataContainer() *spos.ConsensusCore {
	marshalizerMock := mock.MarshalizerMock{}
	blockChain := &testscommon.ChainHandlerStub{}
	blockProcessorMock := consensusMocks.InitBlockProcessorMock(marshalizerMock)
	bootstrapperMock := &bootstrapperStubs.BootstrapperStub{}
	broadcastMessengerMock := &consensusMocks.BroadcastMessengerMock{}
	chronologyHandlerMock := consensusMocks.InitChronologyHandlerMock()
	multiSignerMock := cryptoMocks.NewMultiSigner()
	hasherMock := &hashingMocks.HasherMock{}
	roundHandlerMock := &consensusMocks.RoundHandlerMock{}
	epochStartSubscriber := &epochstartmock.EpochStartNotifierStub{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := &consensusMocks.SyncTimerMock{}
	nodesCoordinator := &shardingMocks.NodesCoordinatorMock{}
	antifloodHandler := &mock.P2PAntifloodHandlerStub{}
	peerHonestyHandler := &testscommon.PeerHonestyHandlerStub{}
	headerSigVerifier := &consensusMocks.HeaderSigVerifierMock{}
	fallbackHeaderValidator := &testscommon.FallBackHeaderValidatorStub{}
	nodeRedundancyHandler := &mock.NodeRedundancyHandlerStub{}
	scheduledProcessor := &consensusMocks.ScheduledProcessorStub{}
	messageSigningHandler := &mock.MessageSigningHandlerStub{}
	peerBlacklistHandler := &mock.PeerBlacklistHandlerStub{}
	multiSignerContainer := cryptoMocks.NewMultiSignerContainerMock(multiSignerMock)
	signingHandler := &consensusMocks.SigningHandlerStub{}
	enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{}
	proofsPool := &dataRetriever.ProofsPoolMock{}
	epochNotifier := &epochNotifierMock.EpochNotifierStub{}
	invalidSignersCache := &consensusMocks.InvalidSignersCacheMock{}

	consensusCore, _ := spos.NewConsensusCore(&spos.ConsensusCoreArgs{
		BlockChain:                    blockChain,
		BlockProcessor:                blockProcessorMock,
		Bootstrapper:                  bootstrapperMock,
		BroadcastMessenger:            broadcastMessengerMock,
		ChronologyHandler:             chronologyHandlerMock,
		Hasher:                        hasherMock,
		Marshalizer:                   marshalizerMock,
		MultiSignerContainer:          multiSignerContainer,
		RoundHandler:                  roundHandlerMock,
		ShardCoordinator:              shardCoordinatorMock,
		SyncTimer:                     syncTimerMock,
		NodesCoordinator:              nodesCoordinator,
		EpochStartRegistrationHandler: epochStartSubscriber,
		AntifloodHandler:              antifloodHandler,
		PeerHonestyHandler:            peerHonestyHandler,
		HeaderSigVerifier:             headerSigVerifier,
		FallbackHeaderValidator:       fallbackHeaderValidator,
		NodeRedundancyHandler:         nodeRedundancyHandler,
		ScheduledProcessor:            scheduledProcessor,
		MessageSigningHandler:         messageSigningHandler,
		PeerBlacklistHandler:          peerBlacklistHandler,
		SigningHandler:                signingHandler,
		EnableEpochsHandler:           enableEpochsHandler,
		EquivalentProofsPool:          proofsPool,
		EpochNotifier:                 epochNotifier,
		InvalidSignersCache:           invalidSignersCache,
	})

	return consensusCore
}

func TestConsensusContainerValidator_ValidateNilConsensusCoreFail(t *testing.T) {
	t.Parallel()

	err := spos.ValidateConsensusCore(nil)

	assert.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestConsensusContainerValidator_ValidateNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetBlockchain(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestConsensusContainerValidator_ValidateNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetBlockProcessor(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestConsensusContainerValidator_ValidateNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetBootStrapper(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestConsensusContainerValidator_ValidateNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetChronology(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetHasher(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestConsensusContainerValidator_ValidateNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetMarshalizer(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetMultiSignerContainer(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetMultiSignerContainer(cryptoMocks.NewMultiSignerContainerMock(nil))

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestConsensusContainerValidator_ValidateNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetRoundHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestConsensusContainerValidator_ValidateNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetShardCoordinator(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestConsensusContainerValidator_ValidateNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetSyncTimer(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestConsensusContainerValidator_ValidateNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetNodesCoordinator(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestConsensusContainerValidator_ValidateNilAntifloodHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetAntifloodHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilAntifloodHandler, err)
}

func TestConsensusContainerValidator_ValidateNilPeerHonestyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetPeerHonestyHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilPeerHonestyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilHeaderSigVerifierShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetHeaderSigVerifier(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilHeaderSigVerifier, err)
}

func TestConsensusContainerValidator_ValidateNilFallbackHeaderValidatorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetFallbackHeaderValidator(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilFallbackHeaderValidator, err)
}

func TestConsensusContainerValidator_ValidateNilNodeRedundancyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetNodeRedundancyHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilNodeRedundancyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilSignatureHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetSigningHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilSigningHandler, err)
}

func TestConsensusContainerValidator_ValidateNilEnableEpochsHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetEnableEpochsHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilEnableEpochsHandler, err)
}

func TestConsensusContainerValidator_ValidateNilBroadcastMessengerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetBroadcastMessenger(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilBroadcastMessenger, err)
}

func TestConsensusContainerValidator_ValidateNilScheduledProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetScheduledProcessor(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilScheduledProcessor, err)
}

func TestConsensusContainerValidator_ValidateNilMessageSigningHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetMessageSigningHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilMessageSigningHandler, err)
}

func TestConsensusContainerValidator_ValidateNilPeerBlacklistHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetPeerBlacklistHandler(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilPeerBlacklistHandler, err)
}

func TestConsensusContainerValidator_ValidateNilEquivalentProofPoolShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetEquivalentProofsPool(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilEquivalentProofPool, err)
}

func TestConsensusContainerValidator_ValidateNilEpochNotifierShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetEpochNotifier(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilEpochNotifier, err)
}

func TestConsensusContainerValidator_ValidateNilEpochStartRegistrationHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetEpochStartNotifier(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilEpochStartNotifier, err)
}

func TestConsensusContainerValidator_ValidateNilInvalidSignersCacheShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.SetInvalidSignersCache(nil)

	err := spos.ValidateConsensusCore(container)

	assert.Equal(t, spos.ErrNilInvalidSignersCache, err)
}

func TestConsensusContainerValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	err := spos.ValidateConsensusCore(container)

	assert.Nil(t, err)
}
