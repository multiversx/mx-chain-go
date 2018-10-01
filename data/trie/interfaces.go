package trie

// DatabaseReader wraps the Get and Has method of a backing store for the trie.
type DatabaseReader interface {
	// Get retrieves the value associated with key from the database.
	Get(key []byte) (value []byte, err error)

	// Has retrieves whether a key is present in the database.
	Has(key []byte) (bool, error)
}

// Code using batches should try to add this much data to the batch.
// The value was determined empirically.
const IdealBatchSize = 100 * 1024

type Persister interface {
	// Add the value to the (key, val) persistance medium
	Putter

	// gets the value associated to the key
	Get(key []byte) ([]byte, error)

	// returns true if the given key is present in the persistance medium
	Has(key []byte) (bool, error)

	// initialized the persistance medium and prepares it for usage
	Init() error

	// Closes the files/resources associated to the persistance medium
	Close() error

	// Deletes the data associated to the given key
	Deleter

	// Removes the persistance medium stored data
	Destroy() error
}

// Database wraps all database operations. All methods are safe for concurrent use.
type PersistDB interface {
	Persister
	NewBatch() Batch
}

// Putter wraps the database write operation supported by both batches and regular databases.
type Putter interface {
	Put(key []byte, value []byte) error
}

// Deleter wraps the database delete operation supported by both batches and regular databases.
type Deleter interface {
	Delete(key []byte) error
}

// Batch is a write-only database that commits changes to its host database
// when Write is called. Batch cannot be used concurrently.
type Batch interface {
	Putter
	Deleter
	ValueSize() int // amount of data in the batch
	Write() error
	// Reset resets the batch for reuse
	Reset()
}
