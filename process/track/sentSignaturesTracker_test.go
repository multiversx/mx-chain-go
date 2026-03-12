package track

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon"
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

func TestSentSignaturesTracker_ResetCountersForManagedBlockSigner(t *testing.T) {
	t.Parallel()

	pk1 := []byte("pk1")
	pk2 := []byte("pk2")

	t.Run("empty map should call remove", func(t *testing.T) {
		t.Parallel()

		pkBytesSlice := make([][]byte, 0)
		keysHandler := &testscommon.KeysHandlerStub{
			ResetRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte, pid core.PeerID) {
				assert.Equal(t, externalPeerID.Bytes(), pid.Bytes())
				pkBytesSlice = append(pkBytesSlice, pkBytes)
			},
		}

		tracker, _ := NewSentSignaturesTracker(keysHandler)
		tracker.ResetCountersForManagedBlockSigner(pk1)

		assert.Equal(t, [][]byte{pk1}, pkBytesSlice)
	})
	t.Run("should call remove only for the public key that did not sent signatures from self", func(t *testing.T) {
		t.Parallel()

		pkBytesSlice := make([][]byte, 0)
		keysHandler := &testscommon.KeysHandlerStub{
			ResetRoundsWithoutReceivedMessagesCalled: func(pkBytes []byte, pid core.PeerID) {
				assert.Equal(t, externalPeerID.Bytes(), pid.Bytes())
				pkBytesSlice = append(pkBytesSlice, pkBytes)
			},
		}

		tracker, _ := NewSentSignaturesTracker(keysHandler)
		tracker.SignatureSent(pk1)

		tracker.ResetCountersForManagedBlockSigner(pk1)
		tracker.ResetCountersForManagedBlockSigner(pk2)
		assert.Equal(t, [][]byte{pk2}, pkBytesSlice)

		t.Run("after reset, all should be called", func(t *testing.T) {
			tracker.StartRound()

			tracker.ResetCountersForManagedBlockSigner(pk1)
			tracker.ResetCountersForManagedBlockSigner(pk2)
			assert.Equal(t, [][]byte{
				pk2,      // from the previous test
				pk1, pk2, // from this call
			}, pkBytesSlice)
		})
	})
}

func TestSentSignaturesTracker_RecordSignedNonceAndGetSignedHash(t *testing.T) {
	t.Parallel()

	t.Run("GetSignedHash returns false when no entry exists", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		hash, exists := tracker.GetSignedHash([]byte("pk1"), 100)
		assert.Nil(t, hash)
		assert.False(t, exists)
	})

	t.Run("RecordSignedNonce stores entry and GetSignedHash retrieves it", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		pk := []byte("pk1")
		nonce := uint64(100)
		headerHash := []byte("hash_A")

		tracker.RecordSignedNonce(pk, nonce, headerHash)
		hash, exists := tracker.GetSignedHash(pk, nonce)

		require.True(t, exists)
		assert.Equal(t, headerHash, hash)
	})

	t.Run("first-write-wins: second write for same pk+nonce is ignored", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		pk := []byte("pk1")
		nonce := uint64(100)
		firstHash := []byte("hash_A")
		secondHash := []byte("hash_B")

		tracker.RecordSignedNonce(pk, nonce, firstHash)
		tracker.RecordSignedNonce(pk, nonce, secondHash)

		hash, exists := tracker.GetSignedHash(pk, nonce)
		require.True(t, exists)
		assert.Equal(t, firstHash, hash, "first-write-wins: should return the first hash")
	})

	t.Run("different pks are independent", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		nonce := uint64(100)

		tracker.RecordSignedNonce([]byte("pk1"), nonce, []byte("hash_A"))
		tracker.RecordSignedNonce([]byte("pk2"), nonce, []byte("hash_B"))

		hash1, exists1 := tracker.GetSignedHash([]byte("pk1"), nonce)
		hash2, exists2 := tracker.GetSignedHash([]byte("pk2"), nonce)

		require.True(t, exists1)
		require.True(t, exists2)
		assert.Equal(t, []byte("hash_A"), hash1)
		assert.Equal(t, []byte("hash_B"), hash2)
	})

	t.Run("cleanup removes entries more than 1 nonce behind", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		pk := []byte("pk1")

		tracker.RecordSignedNonce(pk, 100, []byte("hash_100"))
		tracker.RecordSignedNonce(pk, 101, []byte("hash_101"))

		// Both should exist (101 - 100 = 1, not > 1)
		_, exists100 := tracker.GetSignedHash(pk, 100)
		_, exists101 := tracker.GetSignedHash(pk, 101)
		assert.True(t, exists100, "nonce 100 should still exist (1 behind)")
		assert.True(t, exists101)

		// Recording nonce 102 should clean up nonce 100 (102 - 100 = 2 > 1)
		tracker.RecordSignedNonce(pk, 102, []byte("hash_102"))

		_, exists100 = tracker.GetSignedHash(pk, 100)
		_, exists101 = tracker.GetSignedHash(pk, 101)
		_, exists102 := tracker.GetSignedHash(pk, 102)

		assert.False(t, exists100, "nonce 100 should be cleaned up (2 behind)")
		assert.True(t, exists101, "nonce 101 should still exist (1 behind)")
		assert.True(t, exists102)
	})

	t.Run("cleanup does not affect other pks", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})

		tracker.RecordSignedNonce([]byte("pk1"), 100, []byte("hash_pk1_100"))
		tracker.RecordSignedNonce([]byte("pk2"), 100, []byte("hash_pk2_100"))

		// Advancing pk1 to nonce 102 should clean up pk1's nonce 100 but not pk2's
		tracker.RecordSignedNonce([]byte("pk1"), 102, []byte("hash_pk1_102"))

		_, pk1Exists := tracker.GetSignedHash([]byte("pk1"), 100)
		_, pk2Exists := tracker.GetSignedHash([]byte("pk2"), 100)

		assert.False(t, pk1Exists, "pk1 nonce 100 should be cleaned up")
		assert.True(t, pk2Exists, "pk2 nonce 100 should NOT be cleaned up by pk1's advancement")
	})

	t.Run("stores a copy of the header hash, not a reference", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		pk := []byte("pk1")
		nonce := uint64(100)
		headerHash := []byte("hash_A")

		tracker.RecordSignedNonce(pk, nonce, headerHash)

		// Mutate the original slice
		headerHash[0] = 'X'

		hash, exists := tracker.GetSignedHash(pk, nonce)
		require.True(t, exists)
		assert.Equal(t, []byte("hash_A"), hash, "stored hash should not be affected by mutation of the original")
	})

	t.Run("StartRound does not clear signedNonces", func(t *testing.T) {
		t.Parallel()

		tracker, _ := NewSentSignaturesTracker(&testscommon.KeysHandlerStub{})
		pk := []byte("pk1")
		nonce := uint64(100)

		tracker.RecordSignedNonce(pk, nonce, []byte("hash_A"))
		tracker.StartRound()

		hash, exists := tracker.GetSignedHash(pk, nonce)
		require.True(t, exists, "signedNonces should persist across StartRound")
		assert.Equal(t, []byte("hash_A"), hash)
	})
}
