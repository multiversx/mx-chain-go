package antiflood_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood"
	"github.com/multiversx/mx-chain-go/process/throttle/antiflood/disabled"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func TestNewP2PAntiflood_NilBlacklistHandlerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		nil,
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrNilBlackListCacher))
}

func TestNewP2PAntiflood_EmptyFloodPreventerListShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrEmptyFloodPreventerList))
}

func TestNewP2PAntiflood_NilTopicFloodPreventerShouldErr(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		nil,
		&mock.FloodPreventerStub{},
	)
	assert.True(t, check.IfNil(afm))
	assert.True(t, errors.Is(err, process.ErrNilTopicFloodPreventer))
}

func TestNewP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	assert.False(t, check.IfNil(afm))
	assert.Nil(t, err)
}

//------- CanProcessMessage

func TestP2PAntiflood_CanProcessMessageNilMessageShouldError(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessage(nil, "connected peer")
	assert.Equal(t, p2p.ErrNilMessage, err)
}

func TestP2PAntiflood_CanNotIncrementFromConnectedPeerShouldError(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := core.PeerID("from connected peer")
	message := &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
		FromField: messageOriginator,
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, size uint64) error {
				if pid != fromConnectedPeer {
					assert.Fail(t, "should have been the connected peer")
				}

				return process.ErrSystemBusy
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2PAntiflood_CanNotIncrementMessageOriginatorShouldError(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := core.PeerID("from connected peer")
	message := &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
		FromField: messageOriginator,
		PeerField: core.PeerID(messageOriginator),
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, size uint64) error {
				if pid == message.PeerField {
					return process.ErrSystemBusy
				}
				if pid != fromConnectedPeer {
					return process.ErrSystemBusy
				}

				return nil
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2PAntiflood_ShouldWork(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := core.PeerID("from connected peer")
	message := &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: core.PeerID(messageOriginator),
	}
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, size uint64) error {
				return nil
			},
		},
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
}

func TestP2PAntiflood_ShouldWorkWithMoreThanOneFlodPreventer(t *testing.T) {
	t.Parallel()

	messageOriginator := []byte("originator")
	fromConnectedPeer := core.PeerID("from connected peer")
	message := &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: core.PeerID(messageOriginator),
	}
	numIncreasedLoads := 0

	fp := &mock.FloodPreventerStub{
		IncreaseLoadCalled: func(pid core.PeerID, size uint64) error {
			numIncreasedLoads++
			return nil
		},
	}

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		fp,
		fp,
	)

	err := afm.CanProcessMessage(message, fromConnectedPeer)
	assert.Nil(t, err)
	assert.Equal(t, 4, numIncreasedLoads)
}

//------- CanProcessMessagesOnTopic

func TestP2pAntiflood_CanProcessMessagesOnTopicCanNotAccumulateShouldError(t *testing.T) {
	t.Parallel()

	numMessagesCall := uint32(78)
	topicCall := "topic"
	identifierCall := core.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(pid core.PeerID, topic string, numMessages uint32) error {
				if pid == identifierCall && topic == topicCall && numMessages == numMessagesCall {
					return process.ErrSystemBusy
				}

				return nil
			},
		},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessagesOnTopic(identifierCall, topicCall, numMessagesCall, 0, nil)

	assert.True(t, errors.Is(err, process.ErrSystemBusy))
}

func TestP2pAntiflood_CanProcessMessagesOnTopicCanAccumulateShouldWork(t *testing.T) {
	t.Parallel()

	numMessagesCall := uint32(78)
	topicCall := "topic"
	identifierCall := core.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			IncreaseLoadCalled: func(pid core.PeerID, topic string, numMessages uint32) error {
				if pid == identifierCall && topic == topicCall && numMessages == numMessagesCall {
					return nil
				}

				return process.ErrSystemBusy
			},
		},
		&mock.FloodPreventerStub{},
	)

	err := afm.CanProcessMessagesOnTopic(identifierCall, topicCall, numMessagesCall, 0, nil)

	assert.Nil(t, err)
}

func TestP2pAntiflood_CanProcessMessagesOriginatorIsBlacklistedShouldErr(t *testing.T) {
	t.Parallel()

	identifier := core.PeerID("id")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{
			HasCalled: func(pid core.PeerID) bool {
				return true
			},
		},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			IncreaseLoadCalled: func(pid core.PeerID, size uint64) error {
				return nil
			},
		},
	)
	message := &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
		PeerField: identifier,
	}

	err := afm.CanProcessMessage(message, identifier)

	assert.True(t, errors.Is(err, process.ErrOriginatorIsBlacklisted))
}

func TestP2pAntiflood_ResetForTopicSetMaxMessagesShouldWork(t *testing.T) {
	t.Parallel()

	resetTopicCalled := false
	resetTopicParameter := ""
	setMaxMessagesForTopicCalled := false
	setMaxMessagesForTopicParameter1 := ""
	setMaxMessagesForTopicParameter2 := uint32(0)
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{
			ResetForTopicCalled: func(topic string) {
				resetTopicCalled = true
				resetTopicParameter = topic
			},
			SetMaxMessagesForTopicCalled: func(topic string, num uint32) {
				setMaxMessagesForTopicCalled = true
				setMaxMessagesForTopicParameter1 = topic
				setMaxMessagesForTopicParameter2 = num
			},
		},
		&mock.FloodPreventerStub{},
	)

	resetTopic := "reset topic"
	afm.ResetForTopic(resetTopic)
	assert.True(t, resetTopicCalled)
	assert.Equal(t, resetTopic, resetTopicParameter)

	setMaxMessagesForTopic := "set max message for topic"
	setMaxMessagesForTopicNum := uint32(77463)
	afm.SetMaxMessagesForTopic(setMaxMessagesForTopic, setMaxMessagesForTopicNum)
	assert.True(t, setMaxMessagesForTopicCalled)
	assert.Equal(t, setMaxMessagesForTopic, setMaxMessagesForTopicParameter1)
	assert.Equal(t, setMaxMessagesForTopicNum, setMaxMessagesForTopicParameter2)
}

func TestP2pAntiflood_ApplyConsensusSize(t *testing.T) {
	t.Parallel()

	wasCalled := false
	expectedSize := 878264
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{
			ApplyConsensusSizeCalled: func(size int) {
				assert.Equal(t, expectedSize, size)
				wasCalled = true
			},
		},
	)

	afm.ApplyConsensusSize(expectedSize)
	assert.True(t, wasCalled)
}

func TestP2pAntiflood_SetDebuggerNilDebuggerShouldErr(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	err := afm.SetDebugger(nil)
	assert.Equal(t, process.ErrNilDebugger, err)
}

func TestP2pAntiflood_SetDebuggerShouldWork(t *testing.T) {
	t.Parallel()

	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	debugger := &disabled.AntifloodDebugger{}
	err := afm.SetDebugger(debugger)

	assert.Nil(t, err)
	assert.True(t, afm.Debugger() == debugger)
}

func TestP2pAntiflood_Close(t *testing.T) {
	t.Parallel()

	numCalls := int32(0)
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)
	_ = afm.SetDebugger(&mock.AntifloodDebuggerStub{
		CloseCalled: func() error {
			atomic.AddInt32(&numCalls, 1)

			return nil
		},
	})

	err := afm.Close()

	assert.Nil(t, err)
	assert.Equal(t, int32(1), atomic.LoadInt32(&numCalls))
}

func TestP2pAntiflood_BlacklistPeerErrShouldDoNothing(t *testing.T) {
	t.Parallel()

	numCalls := int32(0)
	expectedErr := errors.New("expected error")
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid core.PeerID, span time.Duration) error {
				atomic.AddInt32(&numCalls, 1)

				return expectedErr
			},
		},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	afm.BlacklistPeer("pid", "reason", time.Second)

	assert.Equal(t, int32(1), atomic.LoadInt32(&numCalls))
}

func TestP2pAntiflood_BlacklistPeerShouldWork(t *testing.T) {
	t.Parallel()

	numCalls := int32(0)
	afm, _ := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{
			UpsertCalled: func(pid core.PeerID, span time.Duration) error {
				atomic.AddInt32(&numCalls, 1)

				return nil
			},
		},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	afm.BlacklistPeer("pid", "reason", time.Second)

	assert.Equal(t, int32(1), atomic.LoadInt32(&numCalls))
}

func TestP2pAntiflood_IsOriginatorEligibleForTopic(t *testing.T) {
	t.Parallel()

	afm, err := antiflood.NewP2PAntiflood(
		&mock.PeerBlackListHandlerStub{},
		&mock.TopicAntiFloodStub{},
		&mock.FloodPreventerStub{},
	)

	assert.False(t, check.IfNil(afm))
	assert.Nil(t, err)

	err = afm.SetPeerValidatorMapper(nil)
	assert.Equal(t, err, process.ErrNilPeerValidatorMapper)

	peerValidatorMapper := &mock.PeerShardResolverStub{
		GetPeerInfoCalled: func(pid core.PeerID) core.P2PPeerInfo {
			return core.P2PPeerInfo{PeerType: core.UnknownPeer}
		}}
	err = afm.SetPeerValidatorMapper(peerValidatorMapper)
	assert.Nil(t, err)

	err = afm.IsOriginatorEligibleForTopic("test", "topic")
	assert.Equal(t, err, process.ErrOnlyValidatorsCanUseThisTopic)

	afm.SetTopicsForAll("topicForAll1", "topicForAll2")
	err = afm.IsOriginatorEligibleForTopic("test", "topicForAll1")
	assert.Nil(t, err)

	validatorPID := "validator"
	peerValidatorMapper.GetPeerInfoCalled = func(pid core.PeerID) core.P2PPeerInfo {
		if string(pid) == validatorPID {
			return core.P2PPeerInfo{PeerType: core.ValidatorPeer}
		}
		return core.P2PPeerInfo{PeerType: core.UnknownPeer}
	}

	err = afm.IsOriginatorEligibleForTopic(core.PeerID(validatorPID), "topicForAll1")
	assert.Nil(t, err)
	err = afm.IsOriginatorEligibleForTopic(core.PeerID(validatorPID), "topic")
	assert.Nil(t, err)
}
