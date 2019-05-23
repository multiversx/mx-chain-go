package storage

// Persister provides storage of data services in a database like construct
type Persister interface {
	// Put add the value to the (key, val) persistance medium
	Put(key, val []byte) error
	// Get gets the value associated to the key
	Get(key []byte) ([]byte, error)
	// Has returns true if the given key is present in the persistance medium
	Has(key []byte) error
	// Init initializes the persistance medium and prepares it for usage
	Init() error
	// Close closes the files/resources associated to the persistance medium
	Close() error
	// Remove removes the data associated to the given key
	Remove(key []byte) error
	// Destroy removes the persistance medium stored data
	Destroy() error
}

// Cacher provides caching services
type Cacher interface {
	// Clear is used to completely clear the cache.
	Clear()
	// Put adds a value to the cache.  Returns true if an eviction occurred.
	Put(key []byte, value interface{}) (evicted bool)
	// Get looks up a key's value from the cache.
	Get(key []byte) (value interface{}, ok bool)
	// Has checks if a key is in the cache, without updating the
	// recent-ness or deleting it for being stale.
	Has(key []byte) bool
	// Peek returns the key value (or undefined if not found) without updating
	// the "recently used"-ness of the key.
	Peek(key []byte) (value interface{}, ok bool)
	// HasOrAdd checks if a key is in the cache  without updating the
	// recent-ness or deleting it for being stale,  and if not, adds the value.
	// Returns whether found and whether an eviction occurred.
	HasOrAdd(key []byte, value interface{}) (ok, evicted bool)
	// Remove removes the provided key from the cache.
	Remove(key []byte)
	// RemoveOldest removes the oldest item from the cache.
	RemoveOldest()
	// Keys returns a slice of the keys in the cache, from oldest to newest.
	Keys() [][]byte
	// Len returns the number of items in the cache.
	Len() int
	// RegisterHandler registers a new handler to be called when a new data is added
	RegisterHandler(func(key []byte))
}

// BloomFilter provides services for filtering database requests
type BloomFilter interface {
	//Add adds the value to the bloom filter
	Add([]byte)
	// MayContain checks if the value is in in the set. If it returns 'false',
	//the item is definitely not in the DB
	MayContain([]byte) bool
	//Clear sets all the bits from the filter to 0
	Clear()
}

// Storer provides storage services in a two layered storage construct, where the first layer is
// represented by a cache and second layer by a persitent storage (DB-like)
type Storer interface {
	Put(key, data []byte) error
	Get(key []byte) ([]byte, error)
	Has(key []byte) error
	HasOrAdd(key []byte, value []byte) error
	Remove(key []byte) error
	ClearCache()
	DestroyUnit() error
}
