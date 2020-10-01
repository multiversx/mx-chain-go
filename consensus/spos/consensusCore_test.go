package spos_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/stretchr/testify/assert"
)

func createDefaultConsensusCoreArgs() *spos.ConsensusCoreArgs {
	consensusCoreMock := mock.InitConsensusCore()

	args := &spos.ConsensusCoreArgs{
		BlockChain:                    consensusCoreMock.Blockchain(),
		BlockProcessor:                consensusCoreMock.BlockProcessor(),
		Bootstrapper:                  consensusCoreMock.BootStrapper(),
		BroadcastMessenger:            consensusCoreMock.BroadcastMessenger(),
		ChronologyHandler:             consensusCoreMock.Chronology(),
		Hasher:                        consensusCoreMock.Hasher(),
		Marshalizer:                   consensusCoreMock.Marshalizer(),
		BlsPrivateKey:                 consensusCoreMock.PrivateKey(),
		BlsSingleSigner:               consensusCoreMock.SingleSigner(),
		MultiSigner:                   consensusCoreMock.MultiSigner(),
		Rounder:                       consensusCoreMock.Rounder(),
		ShardCoordinator:              consensusCoreMock.ShardCoordinator(),
		NodesCoordinator:              consensusCoreMock.NodesCoordinator(),
		SyncTimer:                     consensusCoreMock.SyncTimer(),
		EpochStartRegistrationHandler: consensusCoreMock.EpochStartRegistrationHandler(),
		AntifloodHandler:              consensusCoreMock.GetAntiFloodHandler(),
		PeerHonestyHandler:            consensusCoreMock.PeerHonestyHandler(),
		HeaderSigVerifier:             consensusCoreMock.HeaderSigVerifier(),
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

func TestConsensusCore_WithNilBlsPrivateKeyShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.BlsPrivateKey = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsPrivateKey, err)
}

func TestConsensusCore_WithNilBlsSingleSignerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.BlsSingleSigner = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilBlsSingleSigner, err)
}

func TestConsensusCore_WithNilMultiSignerShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.MultiSigner = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilMultiSigner, err)
}

func TestConsensusCore_WithNilRounderShouldFail(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	args.Rounder = nil

	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.Nil(t, consensusCore)
	assert.Equal(t, spos.ErrNilRounder, err)
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

func TestConsensusCore_CreateConsensusCoreShouldWork(t *testing.T) {
	t.Parallel()

	args := createDefaultConsensusCoreArgs()
	consensusCore, err := spos.NewConsensusCore(
		args,
	)

	assert.NotNil(t, consensusCore)
	assert.Nil(t, err)
}
