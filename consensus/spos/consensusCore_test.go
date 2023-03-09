package spos_test

import (
	"testing"

	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/stretchr/testify/assert"
)

func createDefaultConsensusCoreArgs() *spos.ConsensusCoreArgs {
	consensusCoreMock := mock.InitConsensusCore()

	scheduledProcessor := &consensus.ScheduledProcessorStub{}

	args := &spos.ConsensusCoreArgs{
		BlockChain:                    consensusCoreMock.Blockchain(),
		BlockProcessor:                consensusCoreMock.BlockProcessor(),
		Bootstrapper:                  consensusCoreMock.BootStrapper(),
		BroadcastMessenger:            consensusCoreMock.BroadcastMessenger(),
		ChronologyHandler:             consensusCoreMock.Chronology(),
		Hasher:                        consensusCoreMock.Hasher(),
		Marshalizer:                   consensusCoreMock.Marshalizer(),
		MultiSignerContainer:          consensusCoreMock.MultiSignerContainer(),
		RoundHandler:                  consensusCoreMock.RoundHandler(),
		ShardCoordinator:              consensusCoreMock.ShardCoordinator(),
		NodesCoordinator:              consensusCoreMock.NodesCoordinator(),
		SyncTimer:                     consensusCoreMock.SyncTimer(),
		EpochStartRegistrationHandler: consensusCoreMock.EpochStartRegistrationHandler(),
		AntifloodHandler:              consensusCoreMock.GetAntiFloodHandler(),
		PeerHonestyHandler:            consensusCoreMock.PeerHonestyHandler(),
		HeaderSigVerifier:             consensusCoreMock.HeaderSigVerifier(),
		FallbackHeaderValidator:       consensusCoreMock.FallbackHeaderValidator(),
		NodeRedundancyHandler:         consensusCoreMock.NodeRedundancyHandler(),
		ScheduledProcessor:            scheduledProcessor,
		MessageSigningHandler:         consensusCoreMock.MessageSigningHandler(),
		PeerBlacklistHandler:          consensusCoreMock.PeerBlacklistHandler(),
		SigningHandler:                consensusCoreMock.SigningHandler(),
	}
	return args
}

func TestConsensusCore_WithNilBlockchainShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.BlockChain = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockChain, err)
}

func TestConsensusCore_WithNilBlockProcessorShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.BlockProcessor = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlockProcessor, err)
}

func TestConsensusCore_WithNilBootstrapperShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.Bootstrapper = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)
	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBootstrapper, err)
}

func TestConsensusCore_WithNilBroadcastMessengerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.BroadcastMessenger = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBroadcastMessenger, err)
}

func TestConsensusCore_WithNilChronologyShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.ChronologyHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)
	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilChronologyHandler, err)
}

func TestConsensusCore_WithNilHasherShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.Hasher = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilHasher, err)
}

func TestConsensusCore_WithNilMarshalizerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.Marshalizer = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMarshalizer, err)
}

func TestConsensusCore_WithNilMultiSignerContainerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.MultiSignerContainer = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMultiSignerContainer, err)
}

func TestConsensusCore_WithNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.MultiSignerContainer = cryptoMocks.NewMultiSignerContainerMock(nil)

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestConsensusCore_WithNilRoundHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.RoundHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilRoundHandler, err)
}

func TestConsensusCore_WithNilShardCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.ShardCoordinator = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestConsensusCore_WithNilNodesCoordinatorShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.NodesCoordinator = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilNodesCoordinator, err)
}

func TestConsensusCore_WithNilSyncTimerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.SyncTimer = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilSyncTimer, err)
}

func TestConsensusCore_WithNilAntifloodHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.AntifloodHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilAntifloodHandler, err)
}

func TestConsensusCore_WithNilPeerHonestyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.PeerHonestyHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilPeerHonestyHandler, err)
}

func TestConsensusCore_WithNilHeaderSigVerifierShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.HeaderSigVerifier = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilHeaderSigVerifier, err)
}

func TestConsensusCore_WithNilFallbackHeaderValidatorShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.FallbackHeaderValidator = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilFallbackHeaderValidator, err)
}

func TestConsensusCore_WithNilNodeRedundancyHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.NodeRedundancyHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilNodeRedundancyHandler, err)
}

func TestConsensusCore_WithNilScheduledProcessorShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.ScheduledProcessor = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilScheduledProcessor, err)
}

func TestConsensusCore_WithNilMessageSigningHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.MessageSigningHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMessageSigningHandler, err)
}

func TestConsensusCore_WithNilPeerBlacklistHandlerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.PeerBlacklistHandler = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilPeerBlacklistHandler, err)
}

func TestConsensusCore_CreateConsensusCoreShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.NotNil(t, consensusCore)
	assert.Nil(t, err)
}
