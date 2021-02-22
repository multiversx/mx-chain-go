package redundancy_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/redundancy"
	"github.com/ElrondNetwork/elrond-go/redundancy/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeRedundancy_ShouldErrNilMessenger(t *testing.T) {
	t.Parallel()

	nr, err := redundancy.NewNodeRedundancy(0, nil)
	assert.Nil(t, nr)
	assert.Equal(t, redundancy.ErrNilMessenger, err)
}

func TestNewNodeRedundancy_ShouldWork(t *testing.T) {
	t.Parallel()

	nr, err := redundancy.NewNodeRedundancy(0, &mock.MessengerStub{})
	assert.NotNil(t, nr)
	assert.Nil(t, err)
}

func TestIsRedundancyNode_ShouldWork(t *testing.T) {
	t.Parallel()

	nr, _ := redundancy.NewNodeRedundancy(-1, &mock.MessengerStub{})
	assert.True(t, nr.IsRedundancyNode())

	nr, _ = redundancy.NewNodeRedundancy(0, &mock.MessengerStub{})
	assert.False(t, nr.IsRedundancyNode())

	nr, _ = redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
	assert.True(t, nr.IsRedundancyNode())
}

func TestIsMainMachineActive_ShouldWork(t *testing.T) {
	t.Parallel()

	maxRoundsOfInactivityAccepted := redundancy.GetMaxRoundsOfInactivityAccepted()

	nr, _ := redundancy.NewNodeRedundancy(-1, &mock.MessengerStub{})
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.True(t, nr.IsMainMachineActive())

	nr, _ = redundancy.NewNodeRedundancy(0, &mock.MessengerStub{})
	assert.False(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.False(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.False(t, nr.IsMainMachineActive())

	nr, _ = redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted)
	assert.False(t, nr.IsMainMachineActive())

	nr, _ = redundancy.NewNodeRedundancy(2, &mock.MessengerStub{})
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted*2 - 1)
	assert.True(t, nr.IsMainMachineActive())

	nr.SetRoundsOfInactivity(maxRoundsOfInactivityAccepted * 2)
	assert.False(t, nr.IsMainMachineActive())
}

func TestAdjustInactivityIfNeeded_ShouldReturnWhenGivenRoundIndexWasAlreadyChecked(t *testing.T) {
	t.Parallel()

	nr, _ := redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
	selfPubKey := "1"
	consensusPubKeys := []string{"1", "2", "3"}

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 0)
	assert.Equal(t, uint64(0), nr.GetRoundsOfInactivity())
}

func TestAdjustInactivityIfNeeded_ShouldNotAdjustIfSelfPubKeyIsNotContainedInConsensusPubKeys(t *testing.T) {
	t.Parallel()

	nr, _ := redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
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

	nr, _ := redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := 0; i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 1)
		assert.Equal(t, uint64(1), nr.GetRoundsOfInactivity())
	}
}

func TestAdjustInactivityIfNeeded_ShouldAdjustCorrectlyInDifferentRounds(t *testing.T) {
	t.Parallel()

	nr, _ := redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := int64(0); i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, i)
		assert.Equal(t, uint64(i), nr.GetRoundsOfInactivity())
	}
}

func TestResetInactivityIfNeeded_ShouldNotResetIfSelfPubKeyIsNotTheSameWithTheConsensusMsgPubKey(t *testing.T) {
	t.Parallel()

	nr, _ := redundancy.NewNodeRedundancy(1, &mock.MessengerStub{})
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
	nr, _ := redundancy.NewNodeRedundancy(1, messengerMock)

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
	nr, _ := redundancy.NewNodeRedundancy(1, messengerMock)

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, "PeerID_2")
	assert.Equal(t, uint64(0), nr.GetRoundsOfInactivity())
}
