package random

import (
	"crypto/rand"
	"encoding/binary"
)

// ConcurrentSafeIntRandomizer implements dataRetriever.IntRandomizer and can be accessed in a concurrent manner
type ConcurrentSafeIntRandomizer struct {
}

// Intn returns an int in [0, n) interval
func (csir *ConcurrentSafeIntRandomizer) Intn(n int) int {
	if n <= 0 {
		return 0
	}

	buff := make([]byte, 8)
	_, _ = rand.Reader.Read(buff)
	valUint64 := binary.BigEndian.Uint64(buff)

	return int(valUint64 % uint64(n))
}

// IsInterfaceNil returns true if there is no value under the interface
func (csir *ConcurrentSafeIntRandomizer) IsInterfaceNil() bool {
	return csir == nil
}
