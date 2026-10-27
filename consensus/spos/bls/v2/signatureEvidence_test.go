package v2

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
)

const testEvidenceGroupSize = 9

func createTestEvidence(roundIndex int64, nonce uint64, threshold int, numShares int) *roundSignatureEvidence {
	group := make([]string, testEvidenceGroupSize)
	for i := range group {
		group[i] = string(rune('A' + i))
	}

	bitmap := make([]byte, 2)
	shares := make([][]byte, testEvidenceGroupSize)
	for i := 0; i < numShares; i++ {
		shares[i] = []byte{byte(i + 1)}
		bitmap[i/8] |= 1 << (uint16(i) % 8)
	}

	return &roundSignatureEvidence{
		roundIndex:     roundIndex,
		nonce:          nonce,
		headerHash:     []byte("header hash"),
		threshold:      threshold,
		consensusGroup: group,
		bitmap:         bitmap,
		shares:         shares,
		count:          numShares,
	}
}

func createUnsettledProofsPool() *dataRetrieverTests.ProofsPoolMock {
	return &dataRetrieverTests.ProofsPoolMock{
		GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
			return nil, errors.New("proof not found")
		},
	}
}

func createSettledProofsPool(headerHash []byte) *dataRetrieverTests.ProofsPoolMock {
	return &dataRetrieverTests.ProofsPoolMock{
		GetProofByNonceCalled: func(headerNonce uint64, shardID uint32) (data.HeaderProofHandler, error) {
			return &block.HeaderProof{HeaderHash: headerHash, HeaderNonce: headerNonce}, nil
		},
	}
}

func TestNewSignatureEvidenceStore(t *testing.T) {
	t.Parallel()

	store, err := newSignatureEvidenceStore(nil)
	assert.Nil(t, store)
	assert.NotNil(t, err)

	store, err = newSignatureEvidenceStore(createUnsettledProofsPool())
	assert.Nil(t, err)
	assert.False(t, store.IsInterfaceNil())
}

func TestSignatureEvidenceStore_CaptureAndGetPreviousRoundEvidence(t *testing.T) {
	t.Parallel()

	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

	_, ok := store.GetPreviousRoundEvidence(5, 11)
	assert.False(t, ok)

	ev := createTestEvidence(10, 5, 7, 3)
	store.Capture(ev)

	got, ok := store.GetPreviousRoundEvidence(5, 11)
	require.True(t, ok)
	assert.Equal(t, ev, got)

	_, ok = store.GetPreviousRoundEvidence(6, 11)
	assert.False(t, ok, "nonce mismatch should not match")

	_, ok = store.GetPreviousRoundEvidence(5, 12)
	assert.False(t, ok, "evidence older than one round should not match")

	otherEv := createTestEvidence(11, 6, 7, 3)
	store.Capture(otherEv)

	_, ok = store.GetPreviousRoundEvidence(5, 11)
	assert.False(t, ok, "single slot should have been overwritten")

	got, ok = store.GetPreviousRoundEvidence(6, 12)
	require.True(t, ok)
	assert.Equal(t, otherEv, got)
}

func TestSignatureEvidenceStore_CaptureNilClearsSlot(t *testing.T) {
	t.Parallel()

	store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

	subQuorumEv := createTestEvidence(10, 5, 7, 3)
	store.Capture(subQuorumEv)
	store.Capture(nil)

	_, ok := store.GetPreviousRoundEvidence(5, 11)
	assert.False(t, ok)

	_, ok = store.GetRetainedQuorumEvidence(5)
	assert.False(t, ok, "sub-quorum evidence should never be retained")
}

func TestSignatureEvidenceStore_RetentionRule(t *testing.T) {
	t.Parallel()

	t.Run("quorum evidence for unsettled nonce is promoted on overwrite", func(t *testing.T) {
		t.Parallel()

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

		quorumEv := createTestEvidence(10, 5, 7, 8)
		store.Capture(quorumEv)
		store.Capture(nil)

		got, ok := store.GetRetainedQuorumEvidence(5)
		require.True(t, ok)
		assert.Equal(t, quorumEv, got)

		_, ok = store.GetRetainedQuorumEvidence(6)
		assert.False(t, ok)
	})

	t.Run("quorum evidence for settled nonce is dropped", func(t *testing.T) {
		t.Parallel()

		quorumEv := createTestEvidence(10, 5, 7, 8)
		store, _ := newSignatureEvidenceStore(createSettledProofsPool(quorumEv.headerHash))

		store.Capture(quorumEv)
		store.Capture(nil)

		_, ok := store.GetRetainedQuorumEvidence(5)
		assert.False(t, ok)
	})

	t.Run("retained slot is cleared by DropRetained", func(t *testing.T) {
		t.Parallel()

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

		quorumEv := createTestEvidence(10, 5, 7, 8)
		store.Capture(quorumEv)
		store.Capture(nil)

		store.DropRetained(6)
		_, ok := store.GetRetainedQuorumEvidence(5)
		assert.True(t, ok, "drop for another nonce should not clear the slot")

		store.DropRetained(5)
		_, ok = store.GetRetainedQuorumEvidence(5)
		assert.False(t, ok)
	})
}

func TestSignatureEvidenceStore_GetAssemblyCandidate(t *testing.T) {
	t.Parallel()

	t.Run("fresh quorum evidence for unsettled nonce", func(t *testing.T) {
		t.Parallel()

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

		ev := createTestEvidence(10, 5, 7, 8)
		store.Capture(ev)

		got, ok := store.GetAssemblyCandidate()
		require.True(t, ok)
		assert.Equal(t, ev, got)
	})

	t.Run("no candidate for sub-quorum evidence", func(t *testing.T) {
		t.Parallel()

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

		store.Capture(createTestEvidence(10, 5, 7, 6))

		_, ok := store.GetAssemblyCandidate()
		assert.False(t, ok)
	})

	t.Run("no candidate for settled nonce", func(t *testing.T) {
		t.Parallel()

		ev := createTestEvidence(10, 5, 7, 8)
		store, _ := newSignatureEvidenceStore(createSettledProofsPool(ev.headerHash))

		store.Capture(ev)

		_, ok := store.GetAssemblyCandidate()
		assert.False(t, ok)
	})

	t.Run("retained evidence is a candidate while unsettled", func(t *testing.T) {
		t.Parallel()

		store, _ := newSignatureEvidenceStore(createUnsettledProofsPool())

		ev := createTestEvidence(10, 5, 7, 8)
		store.Capture(ev)
		store.Capture(nil)

		got, ok := store.GetAssemblyCandidate()
		require.True(t, ok)
		assert.Equal(t, ev, got)
	})
}

func TestRoundSignatureEvidence_DropShares(t *testing.T) {
	t.Parallel()

	ev := createTestEvidence(10, 5, 7, 8)

	count := ev.dropShares([]int{1, 3})
	assert.Equal(t, 6, count)
	assert.Equal(t, 6, ev.getCount())

	bitmap, shares := ev.getAggregationData()
	assert.Nil(t, shares[1])
	assert.Nil(t, shares[3])
	assert.Equal(t, byte(0), bitmap[0]&(1<<1))
	assert.Equal(t, byte(0), bitmap[0]&(1<<3))
	assert.NotEqual(t, byte(0), bitmap[0]&(1<<2))

	count = ev.dropShares([]int{1, 3, -1, 100})
	assert.Equal(t, 6, count, "repeated or out of range drops should not change the count")
}
