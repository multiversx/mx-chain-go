package random

type intRandomizer interface {
	Intn(n int) int
}

// FisherYatesShuffle will shuffle the provided indexes slice based on a provided randomizer
func FisherYatesShuffle(indexes []int, randomizer intRandomizer) []int {
	newIndexes := make([]int, len(indexes))
	copy(newIndexes, indexes)

	for i := len(newIndexes) - 1; i > 0; i-- {
		j := randomizer.Intn(i + 1)
		newIndexes[i], newIndexes[j] = newIndexes[j], newIndexes[i]
	}

	return newIndexes
}
