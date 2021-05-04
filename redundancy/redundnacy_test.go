package redundancy_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/redundancy"
	"github.com/ElrondNetwork/elrond-go/redundancy/mock"
	"github.com/stretchr/testify/assert"
)

func createMockArguments(redundancyLevel int64) redundancy.ArgNodeRedundancy {
	return redundancy.ArgNodeRedundancy{
		RedundancyLevel:    redundancyLevel,
		Messenger:          &mock.MessengerStub{},
		ObserverPrivateKey: &mock.PrivateKeyStub{},
	}
}

func TestNewNodeRedundancy_ShouldErrNilMessenger(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	arg.Messenger = nil
	nr, err := redundancy.NewNodeRedundancy(arg)

	assert.True(t, check.IfNil(nr))
	assert.Equal(t, redundancy.ErrNilMessenger, err)
}

func TestNewNodeRedundancy_ShouldErrNilObserverPrivateKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	arg.ObserverPrivateKey = nil
	nr, err := redundancy.NewNodeRedundancy(arg)

	assert.True(t, check.IfNil(nr))
	assert.Equal(t, redundancy.ErrNilObserverPrivateKey, err)
}

func TestNewNodeRedundancy_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	nr, err := redundancy.NewNodeRedundancy(arg)

	assert.False(t, check.IfNil(nr))
	assert.Nil(t, err)
}

func TestIsRedundancyNode_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(-1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsRedundancyNode())

	arg = createMockArguments(0)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.False(t, nr.IsRedundancyNode())

	arg = createMockArguments(1)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsRedundancyNode())
}

func TestIsMainMachineActive_ShouldWork(t *testing.T) {
	t.Parallel()

	maxRoundsOfInactivityAccepted := redundancy.GetMaxRoundsOfInactivityAccepted()

	arg := createMockArguments(-1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.True(t, nr.IsMainMachineActive())

	arg = createMockArguments(0)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.False(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.False(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.False(t, nr.IsMainMachineActive())

	arg = createMockArguments(1)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.False(t, nr.IsMainMachineActive())

	arg = createMockArguments(2)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted*2 - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted * 2)
	assert.False(t, nr.IsMainMachineActive())
}

func TestAdjustInactivityIfNeeded_ShouldReturnWhenGivenRoundIndexWasAlreadyChecked(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "1"
	consensusPubKeys := []string{"1", "2", "3"}

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 0)
	assert.Equal(t, uint64(0), nr.GetRoundsOfInactivity())
}

func TestAdjustInactivityIfNeeded_ShouldNotAdjustIfSelfPubKeyIsNotContainedInConsensusPubKeys(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "4"
	consensusPubKeys := []string{"1", "2", "3"}

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 1)
	assert.Equal(t, uint64(0), nr.GetRoundsOfInactivity())

	roundsOfInactivity := redundancy.GetMaxRoundsOfInactivityAccepted()
	nr.SetRoundsOfInactivity(roundsOfInactivity)

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 2)
	assert.Equal(t, roundsOfInactivity, nr.GetRoundsOfInactivity())
}

func TestAdjustInactivityIfNeeded_ShouldAdjustOnlyOneTimeInTheSameRound(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := 0; i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 1)
		assert.Equal(t, uint64(1), nr.GetRoundsOfInactivity())
	}
}

func TestAdjustInactivityIfNeeded_ShouldAdjustCorrectlyInDifferentRounds(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := int64(0); i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, i)
		assert.Equal(t, uint64(i), nr.GetRoundsOfInactivity())
	}
}

func TestResetInactivityIfNeeded_ShouldNotResetIfSelfPubKeyIsNotTheSameWithTheConsensusMsgPubKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "1"
	consensusMsgPubKey := "2"
	consensusMsgPeerID := core.PeerID("PeerID_2")

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	assert.Equal(t, uint64(3), nr.GetRoundsOfInactivity())
}

func TestResetInactivityIfNeeded_ShouldNotResetIfSelfPeerIDIsTheSameWithTheConsensusMsgPeerID(t *testing.T) {
	t.Parallel()

	selfPubKey := "1"
	consensusMsgPubKey := "1"
	consensusMsgPeerID := core.PeerID("PeerID_1")
	messengerMock := &mock.MessengerStub{
		IDCalled: func() core.PeerID {
			return consensusMsgPeerID
		},
	}
	arg := createMockArguments(1)
	arg.Messenger = messengerMock
	nr, _ := redundancy.NewNodeRedundancy(arg)

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	assert.Equal(t, uint64(3), nr.GetRoundsOfInactivity())
}

func TestResetInactivityIfNeeded_ShouldResetRoundsOfInactivity(t *testing.T) {
	t.Parallel()

	selfPubKey := "1"
	consensusMsgPubKey := "1"
	messengerMock := &mock.MessengerStub{
		IDCalled: func() core.PeerID {
			return "PeerID_1"
		},
	}
	arg := createMockArguments(1)
	arg.Messenger = messengerMock
	nr, _ := redundancy.NewNodeRedundancy(arg)

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, "PeerID_2")
	assert.Equal(t, uint64(0), nr.GetRoundsOfInactivity())
}

func TestNodeRedundancy_ObserverPrivateKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(1)
	arg.ObserverPrivateKey = &mock.PrivateKeyStub{}
	nr, _ := redundancy.NewNodeRedundancy(arg)

	assert.True(t, nr.ObserverPrivateKey() == arg.ObserverPrivateKey) //pointer testing
}
