package redundancy_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/redundancy"
	"github.com/multiversx/mx-chain-go/redundancy/mock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
)

func createMockArguments(maxRoundsOfInactivity int) redundancy.ArgNodeRedundancy {
	return redundancy.ArgNodeRedundancy{
		MaxRoundsOfInactivity: maxRoundsOfInactivity,
		Messenger:             &p2pmocks.MessengerStub{},
		ObserverPrivateKey:    &mock.PrivateKeyStub{},
	}
}

func TestNewNodeRedundancy_ShouldErrNilMessenger(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	arg.Messenger = nil
	nr, err := redundancy.NewNodeRedundancy(arg)

	assert.Nil(t, nr)
	assert.Equal(t, redundancy.ErrNilMessenger, err)
}

func TestNewNodeRedundancy_ShouldErrNilObserverPrivateKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	arg.ObserverPrivateKey = nil
	nr, err := redundancy.NewNodeRedundancy(arg)

	assert.Nil(t, nr)
	assert.Equal(t, redundancy.ErrNilObserverPrivateKey, err)
}

func TestNewNodeRedundancy_ShouldErrIfMaxRoundsOfInactivityIsInvalid(t *testing.T) {
	t.Parallel()

	t.Run("maxRoundsOfInactivity is negative", func(t *testing.T) {
		arg := createMockArguments(-1)
		nr, err := redundancy.NewNodeRedundancy(arg)

		assert.Nil(t, nr)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "for maxRoundsOfInactivity, minimum 2 (or 0), got -1")
	})
	t.Run("maxRoundsOfInactivity is 1", func(t *testing.T) {
		arg := createMockArguments(1)
		nr, err := redundancy.NewNodeRedundancy(arg)

		assert.Nil(t, nr)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "for maxRoundsOfInactivity, minimum 2 (or 0), got 1")
	})
}

func TestNewNodeRedundancy_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("maxRoundsOfInactivity is 0", func(t *testing.T) {
		arg := createMockArguments(0)
		nr, err := redundancy.NewNodeRedundancy(arg)

		assert.NotNil(t, nr)
		assert.Nil(t, err)
	})
	t.Run("maxRoundsOfInactivity is 2", func(t *testing.T) {
		arg := createMockArguments(2)
		nr, err := redundancy.NewNodeRedundancy(arg)

		assert.NotNil(t, nr)
		assert.Nil(t, err)
	})
}

func TestNodeRedundancy_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	instance := redundancy.NewNilNodeRedundancy()
	assert.True(t, instance.IsInterfaceNil())

	arg := createMockArguments(2)
	instance, _ = redundancy.NewNodeRedundancy(arg)
	assert.False(t, instance.IsInterfaceNil())
}

func TestIsRedundancyNode_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(0)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	assert.False(t, nr.IsRedundancyNode())

	arg = createMockArguments(2)
	nr, _ = redundancy.NewNodeRedundancy(arg)
	assert.True(t, nr.IsRedundancyNode())
}

func TestIsMainMachineActive_ShouldWork(t *testing.T) {
	t.Parallel()

	t.Run("the node is the main machine", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments(0)
		nr, _ := redundancy.NewNodeRedundancy(arg)
		assert.True(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(1)
		assert.True(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(0)
		assert.True(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(2)
		assert.True(t, nr.IsMainMachineActive())
	})
	t.Run("the node is the backup machine", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments(2)
		nr, _ := redundancy.NewNodeRedundancy(arg)
		assert.True(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(1)
		assert.True(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(2)
		assert.False(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(3)
		assert.False(t, nr.IsMainMachineActive())

		nr.SetRoundsOfInactivity(0)
		assert.True(t, nr.IsMainMachineActive())
	})
}

func TestAdjustInactivityIfNeeded_ShouldReturnWhenGivenRoundIndexWasAlreadyChecked(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "1"
	consensusPubKeys := []string{"1", "2", "3"}

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 0)
	assert.Equal(t, 0, nr.GetRoundsOfInactivity())
}

func TestAdjustInactivityIfNeeded_ShouldNotAdjustIfSelfPubKeyIsNotContainedInConsensusPubKeys(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "4"
	consensusPubKeys := []string{"1", "2", "3"}

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 1)
	assert.Equal(t, 0, nr.GetRoundsOfInactivity())

	nr.SetRoundsOfInactivity(arg.MaxRoundsOfInactivity)

	nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 2)
	assert.Equal(t, arg.MaxRoundsOfInactivity, nr.GetRoundsOfInactivity())
}

func TestAdjustInactivityIfNeeded_ShouldAdjustOnlyOneTimeInTheSameRound(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := 0; i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, 1)
		assert.Equal(t, 1, nr.GetRoundsOfInactivity())
	}
}

func TestAdjustInactivityIfNeeded_ShouldAdjustCorrectlyInDifferentRounds(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "3"
	consensusPubKeys := []string{"1", "2", "3"}

	for i := 0; i < 10; i++ {
		nr.AdjustInactivityIfNeeded(selfPubKey, consensusPubKeys, int64(i))
		assert.Equal(t, i, nr.GetRoundsOfInactivity())
	}
}

func TestResetInactivityIfNeeded_ShouldNotResetIfSelfPubKeyIsNotTheSameWithTheConsensusMsgPubKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	nr, _ := redundancy.NewNodeRedundancy(arg)
	selfPubKey := "1"
	consensusMsgPubKey := "2"
	consensusMsgPeerID := core.PeerID("PeerID_2")

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	assert.Equal(t, 3, nr.GetRoundsOfInactivity())
}

func TestResetInactivityIfNeeded_ShouldNotResetIfSelfPeerIDIsTheSameWithTheConsensusMsgPeerID(t *testing.T) {
	t.Parallel()

	selfPubKey := "1"
	consensusMsgPubKey := "1"
	consensusMsgPeerID := core.PeerID("PeerID_1")
	messengerMock := &p2pmocks.MessengerStub{
		IDCalled: func() core.PeerID {
			return consensusMsgPeerID
		},
	}
	arg := createMockArguments(2)
	arg.Messenger = messengerMock
	nr, _ := redundancy.NewNodeRedundancy(arg)

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, consensusMsgPeerID)
	assert.Equal(t, 3, nr.GetRoundsOfInactivity())
}

func TestResetInactivityIfNeeded_ShouldResetRoundsOfInactivity(t *testing.T) {
	t.Parallel()

	selfPubKey := "1"
	consensusMsgPubKey := "1"
	messengerMock := &p2pmocks.MessengerStub{
		IDCalled: func() core.PeerID {
			return "PeerID_1"
		},
	}
	arg := createMockArguments(2)
	arg.Messenger = messengerMock
	nr, _ := redundancy.NewNodeRedundancy(arg)

	nr.SetRoundsOfInactivity(3)

	nr.ResetInactivityIfNeeded(selfPubKey, consensusMsgPubKey, "PeerID_2")
	assert.Equal(t, 0, nr.GetRoundsOfInactivity())
}

func TestNodeRedundancy_ObserverPrivateKey(t *testing.T) {
	t.Parallel()

	arg := createMockArguments(2)
	arg.ObserverPrivateKey = &mock.PrivateKeyStub{}
	nr, _ := redundancy.NewNodeRedundancy(arg)

	assert.True(t, nr.ObserverPrivateKey() == arg.ObserverPrivateKey) //pointer testing
}
