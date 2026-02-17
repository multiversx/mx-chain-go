package ntpsync

import (
	"fmt"
	"sync"
)

type entry struct {
	nonce uint64
	hash  string
}

type ringBuffer struct {
	mut sync.RWMutex

	buf      []entry
	capacity int
	size     int
	index    int
	set      map[string]struct{} // deduplication map
}

func newNonceRingBuffer(cap int) *ringBuffer {
	return &ringBuffer{
		buf:      make([]entry, cap),
		capacity: cap,
		set:      make(map[string]struct{}),
	}
}

// add adds a nonce if it's not already present. If full, it overwrites the oldest.
func (r *ringBuffer) add(nonce uint64, hash string) {
	r.mut.Lock()
	defer r.mut.Unlock()

	key := createEntryKey(nonce, hash)
	if _, exists := r.set[key]; exists {
		return
	}

	// If overwriting the oldest entry, remove it from set
	if r.size == r.capacity {
		oldest := r.buf[r.index]
		delete(r.set, createEntryKey(oldest.nonce, oldest.hash))
	}

	r.buf[r.index] = entry{
		nonce: nonce,
		hash:  hash,
	}
	r.set[key] = struct{}{}

	// update index and size
	r.index = (r.index + 1) % r.capacity
	if r.size < r.capacity {
		r.size++
	}
}

func createEntryKey(nonce uint64, hash string) string {
	return fmt.Sprintf("%d-%s", nonce, hash)
}

// last returns the last N nonces in chronological order
func (r *ringBuffer) last(n int) []uint64 {
	r.mut.RLock()
	defer r.mut.RUnlock()

	if n > r.size {
		n = r.size
	}
	out := make([]uint64, n)

	start := (r.index - n + r.capacity) % r.capacity
	for i := 0; i < n; i++ {
		idx := (start + i) % r.capacity
		out[i] = r.buf[idx].nonce
	}

	return out
}

func (r *ringBuffer) contains(nonce uint64, hash string) bool {
	r.mut.RLock()
	defer r.mut.RUnlock()

	_, exists := r.set[createEntryKey(nonce, hash)]
	return exists
}

// Size returns number of stored nonces
func (r *ringBuffer) len() int {
	r.mut.RLock()
	defer r.mut.RUnlock()

	return r.size
}
