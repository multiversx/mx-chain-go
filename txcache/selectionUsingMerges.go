package txcache

import (
	"sync"
)

type BunchOfTransactions []*WrappedTransaction

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

	mergedBunch := mergeBunchesInParallel(bunches, numJobsForMerging)
	selection := selectUntilReachedGasRequested(mergedBunch, gasRequested)
	return selection
}

func mergeBunchesInParallel(bunches []BunchOfTransactions, numJobs int) BunchOfTransactions {
	jobs := make([]*mergingJob, numJobs)

	for i := 0; i < numJobs; i++ {
		jobs[i] = &mergingJob{
			input: make([]BunchOfTransactions, 0, len(bunches)/numJobs),
		}
	}

	for i, bunch := range bunches {
		jobs[i%numJobs].input = append(jobs[i%numJobs].input, bunch)
	}

	// Run jobs in parallel
	wg := sync.WaitGroup{}

	for i, job := range jobs {
		wg.Add(1)

		go func(job *mergingJob, i int) {
			job.output = mergeBunches(job.input)
			wg.Done()
		}(job, i)
	}

	wg.Wait()

	// Merge the results of the jobs
	outputBunchesOfJobs := make([]BunchOfTransactions, 0, numJobs)

	for _, job := range jobs {
		outputBunchesOfJobs = append(outputBunchesOfJobs, job.output)
	}

	finalMerge := mergeBunches(outputBunchesOfJobs)
	return finalMerge
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

		if a.isTransactionMoreDesirableByProtocol(b) {
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
