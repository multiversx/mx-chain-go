package resolvers

const (
	maxHashArrayReplySize = 1 << 20
)

type hashArrayReply struct {
	data [][]byte
	size int
}

func newHashArrayReply() hashArrayReply {
	return hashArrayReply{
		data: make([][]byte, 0),
	}
}

func (reply *hashArrayReply) add(data []byte) bool {
	if len(reply.data) > 0 && (reply.size >= maxHashArrayReplySize || len(data) > maxHashArrayReplySize-reply.size) {
		return false
	}

	reply.data = append(reply.data, data)
	reply.size += len(data)

	return true
}

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
