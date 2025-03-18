package proofscache

import (
	"sort"
	"sync"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
)

type proofNonceMapping struct {
	headerHash string
	nonce      uint64
}

type proofsCache struct {
	mutProofsCache sync.RWMutex
	proofsByNonce  []*proofNonceMapping
	proofsByHash   map[string]data.HeaderProofHandler
}

func newProofsCache() *proofsCache {
	return &proofsCache{
		mutProofsCache: sync.RWMutex{},
		proofsByNonce:  make([]*proofNonceMapping, 0),
		proofsByHash:   make(map[string]data.HeaderProofHandler),
	}
}

func (pc *proofsCache) getProofByHash(headerHash []byte) (data.HeaderProofHandler, error) {
	pc.mutProofsCache.RLock()
	defer pc.mutProofsCache.RUnlock()

	proof, ok := pc.proofsByHash[string(headerHash)]
	if !ok {
		return nil, ErrMissingProof
	}

	return proof, nil
}

func (pc *proofsCache) addProof(proof data.HeaderProofHandler) {
	if check.IfNil(proof) {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	pc.proofsByNonce = append(pc.proofsByNonce, &proofNonceMapping{
		headerHash: string(proof.GetHeaderHash()),
		nonce:      proof.GetHeaderNonce(),
	})

	sort.Slice(pc.proofsByNonce, func(i, j int) bool {
		return pc.proofsByNonce[i].nonce < pc.proofsByNonce[j].nonce
	})

	pc.proofsByHash[string(proof.GetHeaderHash())] = proof
}

func (pc *proofsCache) cleanupProofsBehindNonce(nonce uint64) {
	if nonce == 0 {
		return
	}

	pc.mutProofsCache.Lock()
	defer pc.mutProofsCache.Unlock()

	proofsByNonce := make([]*proofNonceMapping, 0)

	for _, proofInfo := range pc.proofsByNonce {
		if proofInfo.nonce < nonce {
			delete(pc.proofsByHash, proofInfo.headerHash)
			continue
		}

		proofsByNonce = append(proofsByNonce, proofInfo)
	}

	pc.proofsByNonce = proofsByNonce
}
