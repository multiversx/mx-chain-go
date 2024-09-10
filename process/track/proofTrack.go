package track

import (
	"sync"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/pkg/errors"
)

// ErrMissingEquivalentProof signals that the equivalent proof is missing
var ErrMissingEquivalentProof = errors.New("missing equivalent proof")

// ProofInfo holds the information about a header
type ProofInfo struct {
	HeaderHash []byte
	Nonce      uint64
	data.HeaderProof
}

type proofTracker struct {
	mutNotarizedProofs sync.RWMutex
	notarizedProofs    map[string]*ProofInfo
}

func NewProofTracker() (*proofTracker, error) {
	return &proofTracker{
		notarizedProofs: make(map[string]*ProofInfo),
	}, nil
}

func (pn *proofTracker) AddNotarizedProof(
	headerHash []byte,
	headerProof data.HeaderProof,
	nonce uint64,
) {
	pn.mutNotarizedProofs.Lock()
	defer pn.mutNotarizedProofs.Unlock()

	pn.notarizedProofs[string(headerHash)] = &ProofInfo{
		HeaderHash:  headerHash,
		HeaderProof: headerProof,
		Nonce:       nonce,
	}
}

func (pn *proofTracker) CleanupNotarizedProofsBehindNonce(shardID uint32, nonce uint64) {
	if nonce == 0 {
		return
	}

	// pn.mutNotarizedProofs.Lock()
	// defer pn.mutNotarizedProofs.Unlock()

	// proofsInfo := make([]*ProofInfo, 0)
	// for _, proofInfo := range pn.notarizedProofs {
	// 	if proofInfo.Nonce < nonce {
	// 		continue
	// 	}

	// 	proofsInfo = append(proofsInfo, proofInfo)
	// }

	// if len(proofsInfo) == 0 {
	// 	proofInfo := pn.lastNotarizedProofInfoUnProtected(shardID)
	// 	if proofInfo == nil {
	// 		return
	// 	}

	// 	proofsInfo = append(proofsInfo, proofInfo)
	// }

	// pn.notarizedProofs[shardID] = proofsInfo
}

func (pn *proofTracker) GetNotarizedProof(
	headerHash []byte,
) (data.HeaderProof, error) {
	pn.mutNotarizedProofs.Lock()
	defer pn.mutNotarizedProofs.Unlock()

	proofInfo, ok := pn.notarizedProofs[string(headerHash)]
	if !ok {
		return data.HeaderProof{}, ErrMissingEquivalentProof
	}

	return proofInfo.HeaderProof, nil
}

// func (pn *proofTracker) lastNotarizedProofInfoUnProtected(shardID uint32) *ProofInfo {
// 	notarizedProofsCount := len(pn.notarizedProofs[shardID])
// 	if notarizedProofsCount > 0 {
// 		return pn.notarizedProofs[shardID][notarizedProofsCount-1]
// 	}

// 	return nil
// }

// IsInterfaceNil returns true if there is no value under the interface
func (pn *proofTracker) IsInterfaceNil() bool {
	return pn == nil
}
