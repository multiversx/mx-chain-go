package txcache

import (
	"bytes"
	"sync"
)

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

func mergeBunchesOfTransactionsInParallel(bunches []BunchOfTransactions) BunchOfTransactions {
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
			job.output = mergeBunches(job.input)
			wg.Done()
		}(job)
	}

	wg.Wait()

	// Merge the results of the jobs
	outputBunchesOfJobs := make([]BunchOfTransactions, 0, numJobsForMerging)

	for _, job := range jobs {
		outputBunchesOfJobs = append(outputBunchesOfJobs, job.output)
	}

	return mergeBunches(outputBunchesOfJobs)
}

func mergeBunches(bunches []BunchOfTransactions) BunchOfTransactions {
	if len(bunches) == 0 {
		return make(BunchOfTransactions, 0)
	}
	if len(bunches) == 1 {
		return bunches[0]
	}

	mid := len(bunches) / 2
	left := mergeBunches(bunches[:mid])
	right := mergeBunches(bunches[mid:])
	return mergeTwoBunches(left, right)
}

// Empty bunches are handled.
func mergeTwoBunches(first BunchOfTransactions, second BunchOfTransactions) BunchOfTransactions {
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
	ppuQuotient := transaction.PricePerGasUnitQuotient
	ppuQuotientOther := otherTransaction.PricePerGasUnitQuotient
	if ppuQuotient != ppuQuotientOther {
		return ppuQuotient > ppuQuotientOther
	}

	ppuRemainder := transaction.PricePerGasUnitRemainder
	ppuRemainderOther := otherTransaction.PricePerGasUnitRemainder
	if ppuRemainder != ppuRemainderOther {
		return ppuRemainder > ppuRemainderOther
	}

	// Then, compare by gas price (to promote the practice of a higher gas price)
	gasPrice := transaction.Tx.GetGasPrice()
	gasPriceOther := otherTransaction.Tx.GetGasPrice()
	if gasPrice != gasPriceOther {
		return gasPrice > gasPriceOther
	}

	// Then, compare by gas limit (promote the practice of lower gas limit)
	// Compare Gas Limits (promote lower gas limit)
	gasLimit := transaction.Tx.GetGasLimit()
	gasLimitOther := otherTransaction.Tx.GetGasLimit()
	if gasLimit != gasLimitOther {
		return gasLimit < gasLimitOther
	}

	// In the end, compare by transaction hash
	return bytes.Compare(transaction.TxHash, otherTransaction.TxHash) > 0
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
