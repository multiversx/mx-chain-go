package immunitycache

// CacheItem defines the interface of a cache item
type CacheItem interface {
	GetKey() []byte
	Payload() interface{}
	Size() int
	IsImmuneToEviction() bool
	ImmunizeAgainstEviction() bool
}

// ForEachItem is an iterator callback
type ForEachItem func(key []byte, value CacheItem)
