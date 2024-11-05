package txcache

type TransactionsMinHeap []*TransactionsHeapItem

func (minHeap TransactionsMinHeap) Len() int { return len(minHeap) }

func (minHeap TransactionsMinHeap) Less(i, j int) bool {
	return minHeap[j].transaction.isTransactionMoreDesirableToNetwork(minHeap[i].transaction)
}

func (minHeap TransactionsMinHeap) Swap(i, j int) {
	minHeap[i], minHeap[j] = minHeap[j], minHeap[i]
}

func (minHeap *TransactionsMinHeap) Push(x interface{}) {
	*minHeap = append(*minHeap, x.(*TransactionsHeapItem))
}

func (minHeap *TransactionsMinHeap) Pop() interface{} {
	// Standard code when storing the heap in a slice:
	// https://pkg.go.dev/container/heap
	old := *minHeap
	n := len(old)
	item := old[n-1]
	*minHeap = old[0 : n-1]
	return item
}
