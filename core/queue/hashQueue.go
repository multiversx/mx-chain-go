package queue

type hashQueue struct {
	queue [][]byte
	size  uint
}

// NewHashQueue creates a new hashQueue
func NewHashQueue(numHashes uint) *hashQueue {
	return &hashQueue{
		queue: make([][]byte, 0),
		size:  numHashes,
	}
}

// Add adds a new element to the queue, and evicts the first element if the queue is full
func (hq *hashQueue) Add(hash []byte) []byte {
	if hq.size == 0 {
		return hash
	}

	if uint(len(hq.queue)) == hq.size {
		hashToEvict := hq.queue[0]
		hq.queue = hq.queue[1:]
		hq.queue = append(hq.queue, hash)

		return hashToEvict
	}

	hq.queue = append(hq.queue, hash)

	return nil
}
