package chainSimulator

import (
	"encoding/binary"
	"sync/atomic"
)

type sequenceGenerator struct {
	cnt uint64
}

// NewSequenceGenerator -
func NewSequenceGenerator() *sequenceGenerator {
	return &sequenceGenerator{
		cnt: 0,
	}
}

// GenerateSequence -
func (generator *sequenceGenerator) GenerateSequence() []byte {
	seqno := make([]byte, 8)
	counter := atomic.AddUint64(&generator.cnt, 1)
	binary.BigEndian.PutUint64(seqno, counter)
	return seqno
}

// IsInterfaceNil -
func (generator *sequenceGenerator) IsInterfaceNil() bool {
	return generator == nil
}
