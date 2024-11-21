package txcache

type transactionsHeap struct {
	items []*transactionsHeapItem
	less  func(i, j int) bool
}

type transactionsHeapItem struct {
	senderIndex      int
	transactionIndex int
	transaction      *WrappedTransaction
}

func newMinTransactionsHeap(capacity int) *transactionsHeap {
	h := transactionsHeap{
		items: make([]*transactionsHeapItem, 0, capacity),
	}

	h.less = func(i, j int) bool {
		return h.items[j].transaction.isTransactionMoreValuableForNetwork(h.items[i].transaction)
	}

	return &h
}

func newMaxTransactionsHeap(capacity int) *transactionsHeap {
	h := transactionsHeap{
		items: make([]*transactionsHeapItem, 0, capacity),
	}

	h.less = func(i, j int) bool {
		return h.items[i].transaction.isTransactionMoreValuableForNetwork(h.items[j].transaction)
	}

	return &h
}

// Len returns the number of elements in the heap.
func (h *transactionsHeap) Len() int { return len(h.items) }

// Less reports whether the element with index i should sort before the element with index j.
func (h *transactionsHeap) Less(i, j int) bool {
	return h.less(i, j)
}

// Swap swaps the elements with indexes i and j.
func (h *transactionsHeap) Swap(i, j int) {
	h.items[i], h.items[j] = h.items[j], h.items[i]
}

// Push pushes the element x onto the heap.
func (h *transactionsHeap) Push(x interface{}) {
	h.items = append(h.items, x.(*transactionsHeapItem))
}

// Pop removes and returns the minimum element (according to "h.less") from the heap.
func (h *transactionsHeap) Pop() interface{} {
	// Standard code when storing the heap in a slice:
	// https://pkg.go.dev/container/heap
	old := h.items
	n := len(old)
	item := old[n-1]
	h.items = old[0 : n-1]
	return item
}
