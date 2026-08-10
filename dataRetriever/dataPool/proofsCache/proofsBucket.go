package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

type proofNonceBucket struct {
	maxNonce      uint64
	proofsByNonce map[uint64]string
}

func newProofBucket() *proofNonceBucket {
	return &proofNonceBucket{
		proofsByNonce: make(map[uint64]string),
	}
}

func (p *proofNonceBucket) size() int {
	return len(p.proofsByNonce)
}

func (p *proofNonceBucket) insert(proof data.HeaderProofHandler) string {
	nonce := proof.GetHeaderNonce()
	newHash := string(proof.GetHeaderHash())

	oldHash, existed := p.proofsByNonce[nonce]
	p.proofsByNonce[nonce] = newHash

	if nonce > p.maxNonce {
		p.maxNonce = nonce
	}

	if existed {
		return oldHash
	}
	return ""
}
