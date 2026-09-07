package resolvers

func deduplicateHashes(hashes [][]byte) [][]byte {
	uniqueHashes := make([][]byte, 0, len(hashes))
	seenHashes := make(map[string]struct{}, len(hashes))

	for _, hash := range hashes {
		hashKey := string(hash)
		if _, alreadySeen := seenHashes[hashKey]; alreadySeen {
			continue
		}

		seenHashes[hashKey] = struct{}{}
		uniqueHashes = append(uniqueHashes, hash)
	}

	return uniqueHashes
}
