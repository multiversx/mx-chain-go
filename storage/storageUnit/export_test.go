package storageUnit

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (s *Unit) GetBlomFilter() storage.BloomFilter {
	return s.bloomFilter
}
