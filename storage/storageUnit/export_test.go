package storageUnit

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

func (s *Unit) GetBlomFilter() storage.BloomFilter {
	return s.bloomFilter
}
