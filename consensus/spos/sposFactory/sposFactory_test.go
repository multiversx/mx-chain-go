package sposFactory_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/mock"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/sposFactory"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/outport"
	statusHandlerMock "github.com/multiversx/mx-chain-go/testscommon/statusHandler"
	"github.com/stretchr/testify/assert"
)

var currentPid = core.PeerID("pid")

func TestGetConsensusCoreFactory_InvalidTypeShouldErr(t *testing.T) {
	t.Parallel()

	csf, err := sposFactory.GetConsensusCoreFactory("invalid")

	assert.Nil(t, csf)
	assert.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}

func TestGetConsensusCoreFactory_BlsShouldWork(t *testing.T) {
	t.Parallel()

	csf, err := sposFactory.GetConsensusCoreFactory(consensus.BlsConsensusType)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(csf))
}

func TestGetSubroundsFactory_BlsNilConsensusCoreShouldErr(t *testing.T) {
	t.Parallel()

	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	statusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	chainID := []byte("chain-id")
	indexer := &outport.OutportStub{}
	sf, err := sposFactory.GetSubroundsFactory(
		nil,
		&spos.ConsensusState{},
		worker,
		consensusType,
		statusHandler,
		indexer,
		chainID,
		currentPid,
	)

	assert.Nil(t, sf)
	assert.Equal(t, spos.ErrNilConsensusCore, err)
}

func TestGetSubroundsFactory_BlsNilStatusHandlerShouldErr(t *testing.T) {
	t.Parallel()

	consensusCore := mock.InitConsensusCore()
	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	chainID := []byte("chain-id")
	indexer := &outport.OutportStub{}
	sf, err := sposFactory.GetSubroundsFactory(
		consensusCore,
		&spos.ConsensusState{},
		worker,
		consensusType,
		nil,
		indexer,
		chainID,
		currentPid,
	)

	assert.Nil(t, sf)
	assert.Equal(t, spos.ErrNilAppStatusHandler, err)
}

func TestGetSubroundsFactory_BlsShouldWork(t *testing.T) {
	t.Parallel()

	consensusCore := mock.InitConsensusCore()
	worker := &mock.SposWorkerMock{}
	consensusType := consensus.BlsConsensusType
	statusHandler := statusHandlerMock.NewAppStatusHandlerMock()
	chainID := []byte("chain-id")
	indexer := &outport.OutportStub{}
	sf, err := sposFactory.GetSubroundsFactory(
		consensusCore,
		&spos.ConsensusState{},
		worker,
		consensusType,
		statusHandler,
		indexer,
		chainID,
		currentPid,
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(sf))
}

func TestGetSubroundsFactory_InvalidConsensusTypeShouldErr(t *testing.T) {
	t.Parallel()

	consensusType := "invalid"
	sf, err := sposFactory.GetSubroundsFactory(
		nil,
		nil,
		nil,
		consensusType,
		nil,
		nil,
		nil,
		currentPid,
	)

	assert.Nil(t, sf)
	assert.Equal(t, sposFactory.ErrInvalidConsensusType, err)
}

func TestGetBroadcastMessenger_ShardShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	messenger := &mock.MessengerStub{}
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.SelfIDCalled = func() uint32 {
		return 0
	}
	privateKey := &mock.PrivateKeyMock{}
	peerSigHandler := &mock.PeerSignatureHandler{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptosContainer := &testscommon.InterceptorsContainerStub{}
	alarmSchedulerStub := &mock.AlarmSchedulerStub{}

	bm, err := sposFactory.GetBroadcastMessenger(
		marshalizer,
		hasher,
		messenger,
		shardCoord,
		privateKey,
		peerSigHandler,
		headersSubscriber,
		interceptosContainer,
		alarmSchedulerStub,
	)

	assert.Nil(t, err)
	assert.NotNil(t, bm)
}

func TestGetBroadcastMessenger_MetachainShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	hasher := &hashingMocks.HasherMock{}
	messenger := &mock.MessengerStub{}
	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.SelfIDCalled = func() uint32 {
		return core.MetachainShardId
	}
	privateKey := &mock.PrivateKeyMock{}
	peerSigHandler := &mock.PeerSignatureHandler{}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptosContainer := &testscommon.InterceptorsContainerStub{}
	alarmSchedulerStub := &mock.AlarmSchedulerStub{}

	bm, err := sposFactory.GetBroadcastMessenger(
		marshalizer,
		hasher,
		messenger,
		shardCoord,
		privateKey,
		peerSigHandler,
		headersSubscriber,
		interceptosContainer,
		alarmSchedulerStub,
	)

	assert.Nil(t, err)
	assert.NotNil(t, bm)
}

func TestGetBroadcastMessenger_NilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	headersSubscriber := &mock.HeadersCacherStub{}
	interceptosContainer := &testscommon.InterceptorsContainerStub{}
	alarmSchedulerStub := &mock.AlarmSchedulerStub{}

	bm, err := sposFactory.GetBroadcastMessenger(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		headersSubscriber,
		interceptosContainer,
		alarmSchedulerStub,
	)

	assert.Nil(t, bm)
	assert.Equal(t, spos.ErrNilShardCoordinator, err)
}

func TestGetBroadcastMessenger_InvalidShardIdShouldErr(t *testing.T) {
	t.Parallel()

	shardCoord := mock.NewMultiShardsCoordinatorMock(3)
	shardCoord.SelfIDCalled = func() uint32 {
		return 37
	}
	headersSubscriber := &mock.HeadersCacherStub{}
	interceptosContainer := &testscommon.InterceptorsContainerStub{}
	alarmSchedulerStub := &mock.AlarmSchedulerStub{}

	bm, err := sposFactory.GetBroadcastMessenger(
		nil,
		nil,
		nil,
		shardCoord,
		nil,
		nil,
		headersSubscriber,
		interceptosContainer,
		alarmSchedulerStub,
	)

	assert.Nil(t, bm)
	assert.Equal(t, sposFactory.ErrInvalidShardId, err)
}
