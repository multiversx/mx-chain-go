package roundSync

import (
	"fmt"
	"sync"
)

type entry struct {
	round uint64
	hash  string
}

type roundRingBuffer struct {
	mut sync.RWMutex

	buf      []entry
	capacity int
	size     int
	index    int
	set      map[string]struct{} // deduplication map
}

func newRoundRingBuffer(cap int) *roundRingBuffer {
	return &roundRingBuffer{
		buf:      make([]entry, cap),
		capacity: cap,
		set:      make(map[string]struct{}),
	}
}

// add adds a round if it's not already present. If full, it overwrites the oldest.
func (r *roundRingBuffer) add(round uint64, hash string) {
	r.mut.Lock()
	defer r.mut.Unlock()

	key := createEntryKey(round, hash)
	if _, exists := r.set[key]; exists {
		return
	}

	// If overwriting the oldest entry, remove it from set
	if r.size == r.capacity {
		oldest := r.buf[r.index]
		delete(r.set, createEntryKey(oldest.round, oldest.hash))
	}

	r.buf[r.index] = entry{
		round: round,
		hash:  hash,
	}
	r.set[key] = struct{}{}

	// update index and size
	r.index = (r.index + 1) % r.capacity
	if r.size < r.capacity {
		r.size++
	}
}

func createEntryKey(round uint64, hash string) string {
	return fmt.Sprintf("%d-%s", round, hash)
}

// last returns the last N rounds in chronological order
func (r *roundRingBuffer) last(n int) []uint64 {
	r.mut.RLock()
	defer r.mut.RUnlock()

	if n > r.size {
		n = r.size
	}
	out := make([]uint64, n)

	start := (r.index - n + r.capacity) % r.capacity
	for i := 0; i < n; i++ {
		idx := (start + i) % r.capacity
		out[i] = r.buf[idx].round
	}

	return out
}

func (r *roundRingBuffer) contains(round uint64, hash string) bool {
	r.mut.RLock()
	defer r.mut.RUnlock()

	_, exists := r.set[createEntryKey(round, hash)]
	return exists
}

// Size returns number of stored rounds
func (r *roundRingBuffer) len() int {
	r.mut.RLock()
	defer r.mut.RUnlock()

	return r.size
}
