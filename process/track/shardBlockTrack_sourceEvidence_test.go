package track

import (
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/stretchr/testify/require"

	dataRetrieverMock "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/pool"
)

const sourceEvidenceCandidateNonce = uint64(6)

func TestShardBlockTrack_SourceMetaVerdictUsesTrackedAndPooledEvidence(t *testing.T) {
	t.Parallel()

	candidateHash := []byte("candidate")
	candidate := &block.MetaBlockV3{Nonce: sourceEvidenceCandidateNonce, Round: 11}
	snapshot := metaAuthoritySnapshot{
		hashes: map[string]struct{}{string(candidateHash): {}},
		valid:  true,
	}

	t.Run("tracker-only foreign continuation rejects pooled source finality", func(t *testing.T) {
		t.Parallel()

		ownChildHash := []byte("ownChild")
		ownChild := &block.MetaBlockV3{Nonce: 7, Round: 12, PrevHash: candidateHash}
		foreignChildHash := []byte("foreignChild")
		foreignChild := &block.MetaBlockV3{Nonce: 7, Round: 13, PrevHash: []byte("foreignParent")}
		foreignGrandChildHash := []byte("foreignGrandChild")
		foreignGrandChild := &block.MetaBlockV3{Nonce: 8, Round: 14, PrevHash: foreignChildHash}

		sbt := newSourceEvidenceTracker(t,
			[]*HeaderInfo{{Hash: ownChildHash, Header: ownChild}},
			[]*HeaderInfo{
				{Hash: foreignChildHash, Header: foreignChild},
				{Hash: foreignGrandChildHash, Header: foreignGrandChild},
			},
		)

		verdict := sbt.sourceMetaVerdict(candidate, candidateHash, sourceEvidenceCandidateNonce, snapshot, false)
		require.Equal(t, sourceMetaDead, verdict)
		require.True(t, sbt.allPendingSourcesDead([]pendingMetaSource{{
			header: candidate,
			hash:   candidateHash,
			nonce:  sourceEvidenceCandidateNonce,
		}}, false))
	})

	t.Run("tracker-only own reconciliation overrides pooled foreign continuation", func(t *testing.T) {
		t.Parallel()

		ownChildHash := []byte("ownChild")
		ownChild := &block.MetaBlockV3{Nonce: 7, Round: 12, PrevHash: candidateHash}
		ownGrandChildHash := []byte("ownGrandChild")
		ownGrandChild := &block.MetaBlockV3{Nonce: 8, Round: 13, PrevHash: ownChildHash}
		foreignChildHash := []byte("foreignChild")
		foreignChild := &block.MetaBlockV3{Nonce: 7, Round: 14, PrevHash: []byte("foreignParent")}
		foreignGrandChildHash := []byte("foreignGrandChild")
		foreignGrandChild := &block.MetaBlockV3{Nonce: 8, Round: 15, PrevHash: foreignChildHash}

		sbt := newSourceEvidenceTracker(t,
			[]*HeaderInfo{
				{Hash: foreignChildHash, Header: foreignChild},
				{Hash: foreignGrandChildHash, Header: foreignGrandChild},
			},
			[]*HeaderInfo{
				{Hash: ownChildHash, Header: ownChild},
				{Hash: ownGrandChildHash, Header: ownGrandChild},
			},
		)

		verdict := sbt.sourceMetaVerdict(candidate, candidateHash, sourceEvidenceCandidateNonce, snapshot, false)
		require.Equal(t, sourceMetaHeldFinal, verdict)
		require.False(t, sbt.allPendingSourcesDead([]pendingMetaSource{{
			header: candidate,
			hash:   candidateHash,
			nonce:  sourceEvidenceCandidateNonce,
		}}, false))
	})
}

func newSourceEvidenceTracker(
	t *testing.T,
	pooledInfos []*HeaderInfo,
	trackedInfos []*HeaderInfo,
) *shardBlockTrack {
	t.Helper()

	pooledByNonce := make(map[uint64][]*HeaderInfo)
	pooledByHash := make(map[string]*HeaderInfo)
	for _, info := range pooledInfos {
		nonce := info.Header.GetNonce()
		pooledByNonce[nonce] = append(pooledByNonce[nonce], info)
		pooledByHash[string(info.Hash)] = info
	}

	headersPool := &pool.HeadersPoolStub{
		GetHeaderByHashCalled: func(hash []byte) (data.HeaderHandler, error) {
			info, ok := pooledByHash[string(hash)]
			if !ok {
				return nil, errors.New("missing header")
			}

			return info.Header, nil
		},
		GetHeaderByNonceAndShardIdCalled: func(nonce uint64, shardID uint32) ([]data.HeaderHandler, [][]byte, error) {
			if shardID != core.MetachainShardId {
				return nil, nil, errors.New("missing headers")
			}
			infos := pooledByNonce[nonce]
			if len(infos) == 0 {
				return nil, nil, errors.New("missing headers")
			}

			headers := make([]data.HeaderHandler, 0, len(infos))
			hashes := make([][]byte, 0, len(infos))
			for _, info := range infos {
				headers = append(headers, info.Header)
				hashes = append(hashes, info.Hash)
			}

			return headers, hashes, nil
		},
	}
	proofsPool := &dataRetrieverMock.ProofsPoolMock{
		HasProofCalled: func(shardID uint32, hash []byte) bool {
			return shardID == core.MetachainShardId && len(hash) > 0
		},
		GetProofsByNonceCalled: func(nonce uint64, shardID uint32) ([]data.HeaderProofHandler, error) {
			if shardID != core.MetachainShardId || nonce != sourceEvidenceCandidateNonce {
				return nil, errors.New("missing proofs")
			}

			return []data.HeaderProofHandler{&block.HeaderProof{
				HeaderHash:    []byte("candidate"),
				HeaderNonce:   sourceEvidenceCandidateNonce,
				HeaderRound:   11,
				HeaderShardId: core.MetachainShardId,
			}}, nil
		},
	}
	view, err := NewMetaFinalityView(ArgsMetaFinalityView{
		HeadersPool: headersPool,
		ProofsPool:  proofsPool,
	})
	require.NoError(t, err)

	trackedByNonce := make(map[uint64][]*HeaderInfo)
	for _, info := range trackedInfos {
		nonce := info.Header.GetNonce()
		trackedByNonce[nonce] = append(trackedByNonce[nonce], info)
	}

	return &shardBlockTrack{
		baseBlockTrack: &baseBlockTrack{
			headersPool: headersPool,
			proofsPool:  proofsPool,
			headers: map[uint32]map[uint64][]*HeaderInfo{
				core.MetachainShardId: trackedByNonce,
			},
		},
		metaFinalityView: view,
	}
}
