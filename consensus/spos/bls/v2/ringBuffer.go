package v2

import "sync"

type RoundRingBuffer struct {
	mut sync.RWMutex

	buf      []uint64
	capacity int
	size     int
	index    int
	set      map[uint64]struct{} // dedupe instant
}

func NewRoundRingBuffer(cap int) *RoundRingBuffer {
	return &RoundRingBuffer{
		buf:      make([]uint64, cap),
		capacity: cap,
		set:      make(map[uint64]struct{}),
	}
}

// Add adds a round if it's not already present. If full, it overwrites the oldest.
func (r *RoundRingBuffer) Add(round uint64) {
	r.mut.Lock()
	defer r.mut.Unlock()

	if _, exists := r.set[round]; exists {
		return
	}

	// If overwriting the oldest entry, remove it from set
	if r.size == r.capacity {
		oldest := r.buf[r.index]
		delete(r.set, oldest)
	}

	r.buf[r.index] = round
	r.set[round] = struct{}{}

	// update index and size
	r.index = (r.index + 1) % r.capacity
	if r.size < r.capacity {
		r.size++
	}
}

// Last returns the last N rounds in chronological order
func (r *RoundRingBuffer) Last(n int) []uint64 {
	r.mut.Lock()
	defer r.mut.Unlock()

	if n > r.size {
		n = r.size
	}
	out := make([]uint64, n)

	start := (r.index - n + r.capacity) % r.capacity
	for i := 0; i < n; i++ {
		idx := (start + i) % r.capacity
		out[i] = r.buf[idx]
	}

	return out
}

func (r *RoundRingBuffer) Contains(round uint64) bool {
	r.mut.Lock()
	defer r.mut.Unlock()

	_, exists := r.set[round]
	return exists
}

// Size returns number of stored rounds
func (r *RoundRingBuffer) Size() int {
	r.mut.Lock()
	defer r.mut.Unlock()

	return r.size
}
