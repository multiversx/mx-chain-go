package txcache

type TransactionsHeapItem struct {
	senderIndex      int
	transactionIndex int
	transaction      *WrappedTransaction
}

type TransactionsMaxHeap []*TransactionsHeapItem
type TransactionsMinHeap []*TransactionsHeapItem

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

func (minHeap TransactionsMinHeap) Len() int { return len(minHeap) }

func (minHeap TransactionsMinHeap) Less(i, j int) bool {
	return minHeap[j].transaction.isTransactionMoreDesirableByProtocol(minHeap[i].transaction)
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
