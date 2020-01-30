package storageUnit

import (
	"github.com/ElrondNetwork/elrond-go/storage"
)

func (u *Unit) GetBlomFilter() storage.BloomFilter {
	return u.bloomFilter
}
