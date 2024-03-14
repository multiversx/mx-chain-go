package spos

import (
	"testing"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	"github.com/stretchr/testify/assert"
)

func initConsensusDataContainer() *ConsensusCore {
	marshalizerMock := mock.MarshalizerMock{}
	blockChain := &testscommon.ChainHandlerStub{}
	blockProcessorMock := mock.InitBlockProcessorMock(marshalizerMock)
	bootstrapperMock := &mock.BootstrapperStub{}
	broadcastMessengerMock := &mock.BroadcastMessengerMock{}
	chronologyHandlerMock := mock.InitChronologyHandlerMock()
	multiSignerMock := cryptoMocks.NewMultiSigner()
	hasherMock := &hashingMocks.HasherMock{}
	roundHandlerMock := &mock.RoundHandlerMock{}
	shardCoordinatorMock := mock.ShardCoordinatorMock{}
	syncTimerMock := &mock.SyncTimerMock{}
	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{}
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

	return &ConsensusCore{
		blockChain:              blockChain,
		blockProcessor:          blockProcessorMock,
		bootstrapper:            bootstrapperMock,
		broadcastMessenger:      broadcastMessengerMock,
		chronologyHandler:       chronologyHandlerMock,
		hasher:                  hasherMock,
		marshalizer:             marshalizerMock,
		multiSignerContainer:    multiSignerContainer,
		roundHandler:            roundHandlerMock,
		shardCoordinator:        shardCoordinatorMock,
		syncTimer:               syncTimerMock,
		nodesCoordinator:        validatorGroupSelector,
		antifloodHandler:        antifloodHandler,
		peerHonestyHandler:      peerHonestyHandler,
		headerSigVerifier:       headerSigVerifier,
		fallbackHeaderValidator: fallbackHeaderValidator,
		nodeRedundancyHandler:   nodeRedundancyHandler,
		scheduledProcessor:      scheduledProcessor,
		messageSigningHandler:   messageSigningHandler,
		peerBlacklistHandler:    peerBlacklistHandler,
		signingHandler:          signingHandler,
		enableEpochsHandler:     enableEpochsHandler,
	}
}

func TestConsensusContainerValidator_ValidateNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.blockChain = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilBlockChain, err)
}

func TestConsensusContainerValidator_ValidateNilProcessorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.blockProcessor = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilBlockProcessor, err)
}

func TestConsensusContainerValidator_ValidateNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.bootstrapper = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilBootstrapper, err)
}

func TestConsensusContainerValidator_ValidateNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.chronologyHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilChronologyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.hasher = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilHasher, err)
}

func TestConsensusContainerValidator_ValidateNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.marshalizer = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilMarshalizer, err)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.multiSignerContainer = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilMultiSignerContainer, err)
}

func TestConsensusContainerValidator_ValidateNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.multiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(nil)

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilMultiSigner, err)
}

func TestConsensusContainerValidator_ValidateNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.roundHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilRoundHandler, err)
}

func TestConsensusContainerValidator_ValidateNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.shardCoordinator = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilShardCoordinator, err)
}

func TestConsensusContainerValidator_ValidateNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.syncTimer = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilSyncTimer, err)
}

func TestConsensusContainerValidator_ValidateNilValidatorGroupSelectorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.nodesCoordinator = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilNodesCoordinator, err)
}

func TestConsensusContainerValidator_ValidateNilAntifloodHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.antifloodHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilAntifloodHandler, err)
}

func TestConsensusContainerValidator_ValidateNilPeerHonestyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.peerHonestyHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilPeerHonestyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilHeaderSigVerifierShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.headerSigVerifier = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilHeaderSigVerifier, err)
}

func TestConsensusContainerValidator_ValidateNilFallbackHeaderValidatorShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.fallbackHeaderValidator = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilFallbackHeaderValidator, err)
}

func TestConsensusContainerValidator_ValidateNilNodeRedundancyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.nodeRedundancyHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilNodeRedundancyHandler, err)
}

func TestConsensusContainerValidator_ValidateNilSignatureHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.signingHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilSigningHandler, err)
}

func TestConsensusContainerValidator_ValidateNilEnableEpochsHandlerShouldFail(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	container.enableEpochsHandler = nil

	err := ValidateConsensusCore(container)

	assert.Equal(t, ErrNilEnableEpochsHandler, err)
}

func TestConsensusContainerValidator_ShouldWork(t *testing.T) {
	t.Parallel()

	container := initConsensusDataContainer()
	err := ValidateConsensusCore(container)

	assert.Nil(t, err)
}
