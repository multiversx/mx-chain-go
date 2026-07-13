package proofscache

type proofNonceBucket struct {
	maxNonce      uint64
	proofsByNonce map[uint64][]string
}

func newProofBucket() *proofNonceBucket {
	return &proofNonceBucket{
		proofsByNonce: make(map[uint64][]string),
	}
}

func (p *proofNonceBucket) size() int {
	size := 0
	for _, hashes := range p.proofsByNonce {
		size += len(hashes)
	}

	return size
}

func (p *proofNonceBucket) hashesAt(nonce uint64) []string {
	return p.proofsByNonce[nonce]
}

// insert adds the hash at the given nonce, keeping any different hashes already stored there
func (p *proofNonceBucket) insert(nonce uint64, hash string) {
	for _, existingHash := range p.proofsByNonce[nonce] {
		if existingHash == hash {
			return
		}
	}

	p.proofsByNonce[nonce] = append(p.proofsByNonce[nonce], hash)

	if nonce > p.maxNonce {
		p.maxNonce = nonce
	}
}

func (p *proofNonceBucket) remove(nonce uint64, hash string) {
	hashes := p.proofsByNonce[nonce]
	for i, existingHash := range hashes {
		if existingHash == hash {
			p.proofsByNonce[nonce] = append(hashes[:i], hashes[i+1:]...)
			return
		}
	}
}
