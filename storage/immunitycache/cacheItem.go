package immunitycache

import (
	"github.com/ElrondNetwork/elrond-go/core/atomic"
)

type cacheItem struct {
	payload  interface{}
	key      string
	size     int
	isImmune atomic.Flag
}

func newCacheItem(payload interface{}, key string, size int) *cacheItem {
	return &cacheItem{
		payload: payload,
		key:     key,
		size:    size,
	}
}

func (item *cacheItem) isImmuneToEviction() bool {
	return item.isImmune.IsSet()
}

func (item *cacheItem) immunizeAgainstEviction() {
	item.isImmune.Set()
}
