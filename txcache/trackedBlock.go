package txcache

type trackedBlock struct {
	nonce                uint64
	hash                 []byte
	rootHash             []byte
	prevHash             []byte
	breadcrumbsByAddress map[string]*accountBreadcrumb
}

func newTrackedBlock(nonce uint64, blockHash []byte, rootHash []byte, prevHash []byte) *trackedBlock {
	return &trackedBlock{
		nonce:                nonce,
		hash:                 blockHash,
		rootHash:             rootHash,
		prevHash:             prevHash,
		breadcrumbsByAddress: make(map[string]*accountBreadcrumb),
	}
}

func (st *trackedBlock) sameNonce(trackedBlock1 *trackedBlock) bool {
	return st.nonce == trackedBlock1.nonce
}
