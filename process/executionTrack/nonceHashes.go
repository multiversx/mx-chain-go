package executionTrack

// nonceHashes this structure is not concurrent safe
type nonceHash struct {
	nonceHashMap map[uint64]string
}

func newNonceHash() *nonceHash {
	return &nonceHash{
		nonceHashMap: make(map[uint64]string),
	}
}

func (nh *nonceHash) addNonceHash(nonce uint64, hash string) {
	nh.nonceHashMap[nonce] = hash
}

func (nh *nonceHash) getHashByNonce(nonce uint64) string {
	return nh.nonceHashMap[nonce]
}

func (nh *nonceHash) removeByNonce(nonce uint64) {
	delete(nh.nonceHashMap, nonce)
}
