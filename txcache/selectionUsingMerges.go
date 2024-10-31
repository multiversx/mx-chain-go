package txcache

import "sync"

type BunchOfTransactions []*WrappedTransaction

const numJobsForMerging = 4

type mergingJob struct {
	input  []BunchOfTransactions
	output BunchOfTransactions
}

func (cache *TxCache) selectTransactionsUsingMerges(gasRequested uint64) BunchOfTransactions {
	senders := cache.getSenders()
	bunches := make([]BunchOfTransactions, 0, len(senders))

	for _, sender := range senders {
		bunches = append(bunches, sender.getTxsWithoutGaps())
	}

	mergedBunch := mergeBunchesOfTransactionsInParallel(bunches)
	selection := selectUntilReachedGasRequested(mergedBunch, gasRequested)
	return selection
}

func selectUntilReachedGasRequested(bunch BunchOfTransactions, gasRequested uint64) BunchOfTransactions {
	accumulatedGas := uint64(0)

	for index, transaction := range bunch {
		accumulatedGas += transaction.Tx.GetGasLimit()

		if accumulatedGas > gasRequested {
			return bunch[0:index]
		}
	}

	return bunch
}

func mergeBunchesOfTransactionsInParallel(bunches []BunchOfTransactions) BunchOfTransactions {
	// If number of bunches is odd, add a phony bunch (to ease pairing logic).
	if len(bunches)%2 == 1 {
		bunches = append(bunches, make(BunchOfTransactions, 0))
	}

	jobs := make([]*mergingJob, numJobsForMerging)

	for i := 0; i < numJobsForMerging; i++ {
		jobs[i] = &mergingJob{
			input: make([]BunchOfTransactions, 0, len(bunches)/numJobsForMerging),
		}
	}

	for i, bunch := range bunches {
		jobs[i%numJobsForMerging].input = append(jobs[i%numJobsForMerging].input, bunch)
	}

	// Run jobs in parallel
	wg := sync.WaitGroup{}

	for _, job := range jobs {
		wg.Add(1)

		go func(job *mergingJob) {
			job.output = mergeBunchesOfTransactions(job.input)[0]
			defer wg.Done()
		}(job)
	}

	wg.Wait()

	// Merge the results of the jobs
	outputBunchesOfJobs := make([]BunchOfTransactions, 0, numJobsForMerging)

	for _, job := range jobs {
		outputBunchesOfJobs = append(outputBunchesOfJobs, job.output)
	}

	return mergeBunchesOfTransactions(outputBunchesOfJobs)[0]
}

func mergeBunchesOfTransactions(bunches []BunchOfTransactions) []BunchOfTransactions {
	if len(bunches) == 1 {
		return bunches
	}

	// Make pairs of bunches, merge a pair into one bunch.
	newBunches := make([]BunchOfTransactions, 0, len(bunches)/2)

	for i := 0; i < len(bunches); i += 2 {
		first := bunches[i]
		second := bunches[i+1]

		newBunch := mergeTwoBunchesOfTransactions(first, second)
		newBunches = append(newBunches, newBunch)
	}

	// Recursive call:
	return mergeBunchesOfTransactions(newBunches)
}

func mergeTwoBunchesOfTransactions(first BunchOfTransactions, second BunchOfTransactions) BunchOfTransactions {
	result := make(BunchOfTransactions, 0, len(first)+len(second))

	firstIndex := 0
	secondIndex := 0

	for firstIndex < len(first) && secondIndex < len(second) {
		a := first[firstIndex]
		b := second[secondIndex]

		if isTransactionGreater(a, b) {
			result = append(result, a)
			firstIndex++
		} else {
			result = append(result, b)
			secondIndex++
		}
	}

	// Append any remaining elements.
	result = append(result, first[firstIndex:]...)
	result = append(result, second[secondIndex:]...)

	return result
}

// Equality is out of scope (not possible in our case).
func isTransactionGreater(transaction *WrappedTransaction, otherTransaction *WrappedTransaction) bool {
	// First, compare by price per unit
	if transaction.PricePerGasUnitQuotient > otherTransaction.PricePerGasUnitQuotient {
		return true
	}
	if transaction.PricePerGasUnitQuotient < otherTransaction.PricePerGasUnitQuotient {
		return false
	}
	if transaction.PricePerGasUnitRemainder > otherTransaction.PricePerGasUnitRemainder {
		return true
	}
	if transaction.PricePerGasUnitRemainder < otherTransaction.PricePerGasUnitRemainder {
		return false
	}

	// Then, compare by gas price (to promote the practice of a higher gas price)
	if transaction.Tx.GetGasPrice() > otherTransaction.Tx.GetGasPrice() {
		return true
	}
	if transaction.Tx.GetGasPrice() < otherTransaction.Tx.GetGasPrice() {
		return false
	}

	// Then, compare by gas limit (promote the practice of lower gas limit)
	if transaction.Tx.GetGasLimit() < otherTransaction.Tx.GetGasLimit() {
		return true
	}
	if transaction.Tx.GetGasLimit() > otherTransaction.Tx.GetGasLimit() {
		return false
	}

	// In the end, compare by transaction hash
	return string(transaction.TxHash) > string(otherTransaction.TxHash)
}
