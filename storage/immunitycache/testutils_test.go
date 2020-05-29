package immunitycache

import (
	"github.com/ElrondNetwork/elrond-go/core/atomic"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ storage.CacheItem = (*cacheItem)(nil)

type cacheItem struct {
	key      string
	payload  interface{}
	size     int
	isImmune atomic.Flag
}

func newCacheItem(key string) *cacheItem {
	return &cacheItem{
		key:     key,
		payload: emptyStruct,
		size:    42,
	}
}

func newCacheItemWithSize(key string, size int) *cacheItem {
	return &cacheItem{
		key:     key,
		payload: emptyStruct,
		size:    size,
	}
}

func newCacheItemWithPayload(key string, payload interface{}) *cacheItem {
	return &cacheItem{
		key:     key,
		payload: payload,
		size:    42,
	}
}

func (item *cacheItem) GetKey() []byte {
	return []byte(item.key)
}

func (item *cacheItem) Payload() interface{} {
	return item.payload
}

func (item *cacheItem) Size() int {
	return item.size
}

func (item *cacheItem) IsImmuneToEviction() bool {
	return item.isImmune.IsSet()
}

func (item *cacheItem) ImmunizeAgainstEviction() {
	item.isImmune.Set()
}

func keysAsStrings(keys [][]byte) []string {
	result := make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		result[i] = string(keys[i])
	}

	return result
}

func keysAsBytes(keys []string) [][]byte {
	result := make([][]byte, len(keys))
	for i := 0; i < len(keys); i++ {
		result[i] = []byte(keys[i])
	}

	return result
}
