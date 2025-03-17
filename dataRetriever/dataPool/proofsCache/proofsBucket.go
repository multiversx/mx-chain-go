package proofscache

import "github.com/multiversx/mx-chain-core-go/data"

type proofNonceBucket struct {
	maxNonce      uint64
	proofsByNonce []*proofNonceMapping
	bucketSize    int
}

func newProofBucket(bucketSize int) *proofNonceBucket {
	return &proofNonceBucket{
		proofsByNonce: make([]*proofNonceMapping, 0),
		bucketSize:    bucketSize,
	}
}

func (p *proofNonceBucket) size() int {
	return len(p.proofsByNonce)
}

func (p *proofNonceBucket) isFull() bool {
	return len(p.proofsByNonce) >= p.bucketSize
}

func (p *proofNonceBucket) insertInNew(proof data.HeaderProofHandler) {
	p.insert(proof)
	p.maxNonce = proof.GetHeaderNonce()
}

func (p *proofNonceBucket) insertInExisting(proof data.HeaderProofHandler) {
	p.insert(proof)

	if proof.GetHeaderNonce() > p.maxNonce {
		p.maxNonce = proof.GetHeaderNonce()
	}
}

func (p *proofNonceBucket) insert(proof data.HeaderProofHandler) {
	p.proofsByNonce = append(p.proofsByNonce, &proofNonceMapping{
		headerHash: string(proof.GetHeaderHash()),
		nonce:      proof.GetHeaderNonce(),
	})
}
