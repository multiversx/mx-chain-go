package executionTrack

type nonceHashes struct {
	nonceHashesMap map[uint64]map[string]struct{}
}

func newNonceHashes() *nonceHashes {
	return &nonceHashes{
		nonceHashesMap: make(map[uint64]map[string]struct{}),
	}
}

func (nh *nonceHashes) addNonceHash(nonce uint64, hash string) {
	_, found := nh.nonceHashesMap[nonce]
	if !found {
		nh.nonceHashesMap[nonce] = make(map[string]struct{})
	}

	nh.nonceHashesMap[nonce][hash] = struct{}{}
}

func (nh *nonceHashes) getNonceHashes(nonce uint64) []string {
	hashes := make([]string, 0)

	hashesMap, found := nh.nonceHashesMap[nonce]
	if !found {
		return hashes
	}
	for hash := range hashesMap {
		hashes = append(hashes, hash)
	}

	return hashes
}

func (nh *nonceHashes) removeByNonce(nonce uint64) {
	delete(nh.nonceHashesMap, nonce)
}

func (nh *nonceHashes) removeFromNonceProvidedHash(nonce uint64, hash string) {
	_, found := nh.nonceHashesMap[nonce]
	if !found {
		return
	}

	delete(nh.nonceHashesMap[nonce], hash)
}
