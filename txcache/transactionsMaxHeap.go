package txcache

type TransactionsMaxHeap []*TransactionsHeapItem

func (maxHeap TransactionsMaxHeap) Len() int { return len(maxHeap) }

func (maxHeap TransactionsMaxHeap) Less(i, j int) bool {
	return maxHeap[i].transaction.isTransactionMoreDesirableByProtocol(maxHeap[j].transaction)
}

func (maxHeap TransactionsMaxHeap) Swap(i, j int) {
	maxHeap[i], maxHeap[j] = maxHeap[j], maxHeap[i]
}

func (maxHeap *TransactionsMaxHeap) Push(x interface{}) {
	*maxHeap = append(*maxHeap, x.(*TransactionsHeapItem))
}

func (maxHeap *TransactionsMaxHeap) Pop() interface{} {
	// Standard code when storing the heap in a slice:
	// https://pkg.go.dev/container/heap
	old := *maxHeap
	n := len(old)
	item := old[n-1]
	*maxHeap = old[0 : n-1]
	return item
}
