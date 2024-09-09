package track

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/pkg/errors"
)

// adapt block notarizer to proofs
// 	add a ProofInfo similar to HeaderInfo

var ErrNilNotarizedProofInfoForShard = errors.New("nil notarized proof info for shard")

var ErrNotarizedProofOffsetOutOfBound = errors.New("requested offset of the notarized proof is out of bound")

type proofNotarizer struct {
	mutNotarizedProofs sync.RWMutex
	notarizedProofs    map[uint32][]*ProofInfo
}

func NewProofNotarizer() (*proofNotarizer, error) {
	return &proofNotarizer{
		notarizedProofs: make(map[uint32][]*ProofInfo),
	}, nil
}

func (pn *proofNotarizer) AddNotarizedProof(
	shardID uint32,
	nonce uint64,
	notarizedProof data.HeaderProof,
	notarizedHeaderHash []byte,
) {
	pn.mutNotarizedProofs.Lock()
	defer pn.mutNotarizedProofs.Unlock()

	pn.notarizedProofs[shardID] = append(pn.notarizedProofs[shardID], &ProofInfo{
		HeaderHash:  notarizedHeaderHash,
		HeaderProof: notarizedProof,
		Nonce:       nonce,
	})

	sort.Slice(pn.notarizedProofs[shardID], func(i, j int) bool {
		return pn.notarizedProofs[shardID][i].Nonce < pn.notarizedProofs[shardID][j].Nonce
	})
}

func (pn *proofNotarizer) CleanupNotarizedProofsBehindNonce(shardID uint32, nonce uint64) {
	if nonce == 0 {
		return
	}

	pn.mutNotarizedProofs.Lock()
	defer pn.mutNotarizedProofs.Unlock()

	notarizedProofs, ok := pn.notarizedProofs[shardID]
	if !ok {
		return
	}

	proofsInfo := make([]*ProofInfo, 0)
	for _, proofInfo := range notarizedProofs {
		if proofInfo.Nonce < nonce {
			continue
		}

		proofsInfo = append(proofsInfo, proofInfo)
	}

	if len(proofsInfo) == 0 {
		proofInfo := pn.lastNotarizedProofInfoUnProtected(shardID)
		if proofInfo == nil {
			return
		}

		proofsInfo = append(proofsInfo, proofInfo)
	}

	pn.notarizedProofs[shardID] = proofsInfo
}

func (pn *proofNotarizer) GetNotarizedProof(
	shardID uint32,
	offset uint64,
) (data.HeaderProof, error) {
	pn.mutNotarizedProofs.Lock()
	defer pn.mutNotarizedProofs.Unlock()

	proofsInfo := pn.notarizedProofs[shardID]
	if proofsInfo == nil {
		return data.HeaderProof{}, ErrNilNotarizedProofInfoForShard
	}

	notarizedProofsCount := uint64(len(proofsInfo))
	if notarizedProofsCount <= offset {
		return data.HeaderProof{}, ErrNotarizedProofOffsetOutOfBound

	}

	proofInfo := proofsInfo[notarizedProofsCount-offset-1]

	return proofInfo.HeaderProof, nil
}

func (pn *proofNotarizer) lastNotarizedProofInfoUnProtected(shardID uint32) *ProofInfo {
	notarizedProofsCount := len(pn.notarizedProofs[shardID])
	if notarizedProofsCount > 0 {
		return pn.notarizedProofs[shardID][notarizedProofsCount-1]
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (pn *proofNotarizer) IsInterfaceNil() bool {
	return pn == nil
}
