package random

import (
	"crypto/rand"
	"math/big"
)

// IntRandomizerConcurrentSafe implements dataRetriever.IntRandomizer and can be accessed in a concurrent manner
type IntRandomizerConcurrentSafe struct {
}

// Intn returns an int in [0, n) interval
func (ir *IntRandomizerConcurrentSafe) Intn(n int) (int, error) {
	val, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}

	return int(val.Int64()), nil
}
