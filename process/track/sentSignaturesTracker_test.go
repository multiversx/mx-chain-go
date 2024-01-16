package track

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewSentSignaturesTracker(t *testing.T) {
	t.Parallel()

	t.Run("nil keys handler should error", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSentSignaturesTracker(nil)
		assert.Nil(t, tracker)
		assert.Equal(t, ErrNilKeysHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		tracker, err := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		assert.NotNil(t, tracker)
		assert.Nil(t, err)
	})
}

func TestSentSignaturesTracker_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var tracker *sentSignaturesTracker
	assert.True(t, tracker.IsInterfaceNil())

	tracker, _ = NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
	assert.False(t, tracker.IsInterfaceNil())
}

func TestSentSignaturesTracker_ResetCountersManagedBlockSigners(t *testing.T) {
	t.Parallel()

	pk1 := []byte("pk1")
	pk2 := []byte("pk2")
	pk3 := []byte("pk3")
	pk4 := []byte("pk4")

	t.Run("empty map should call remove", func(t *testing.T) {
		t.Parallel()

		pkBytesSlice := make([][]byte, 0)
		keysHandler := &testscommon.KeysHandlerStub{
			ResetRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte, pid core.PeerID) {
				assert.Equal(t, externalPeerID.Bytes(), pid.Bytes())
				pkBytesSlice = append(pkBytesSlice, pkBytes)
			},
		}

		signers := [][]byte{pk1, pk2}
		tracker, _ := NewSentSignaturesTracker(keysHandler)
		tracker.ResetCountersManagedBlockSigners(signers)

		assert.Equal(t, [][]byte{pk1, pk2}, pkBytesSlice)
	})
	t.Run("should call remove only for the public keys that did not sent signatures from self", func(t *testing.T) {
		t.Parallel()

		pkBytesSlice := make([][]byte, 0)
		keysHandler := &testscommon.KeysHandlerStub{
			ResetRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte, pid core.PeerID) {
				assert.Equal(t, externalPeerID.Bytes(), pid.Bytes())
				pkBytesSlice = append(pkBytesSlice, pkBytes)
			},
		}

		signers := [][]byte{pk1, pk2, pk3, pk4}
		tracker, _ := NewSentSignaturesTracker(keysHandler)
		tracker.SignatureSent(pk1)
		tracker.SignatureSent(pk3)

		tracker.ResetCountersManagedBlockSigners(signers)
		assert.Equal(t, [][]byte{pk2, pk4}, pkBytesSlice)

		t.Run("after reset, all should be called", func(t *testing.T) {
			tracker.StartRound()

			tracker.ResetCountersManagedBlockSigners(signers)
			assert.Equal(t, [][]byte{
				pk2, pk4, // from the previous test
				pk1, pk2, pk3, pk4, // from this call
			}, pkBytesSlice)
		})
	})
}
