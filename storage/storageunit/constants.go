package storageunit

import "github.com/multiversx/mx-chain-storage-go/storageUnit"

const (
	// LRUCache defines a cache identifier with least-recently-used eviction mechanism
	LRUCache = storageUnit.LRUCache
	// SizeLRUCache defines a cache identifier with least-recently-used eviction mechanism and fixed size in bytes
	SizeLRUCache = storageUnit.SizeLRUCache
)

// LvlDB currently the only supported DBs
// More to be added
const (
	// LvlDB represents a levelDB storage identifier
	LvlDB = storageUnit.LvlDB
	// LvlDBSerial represents a levelDB storage with serialized operations identifier
	LvlDBSerial = storageUnit.LvlDBSerial
	// MemoryDB represents an in memory storage identifier
	MemoryDB = storageUnit.MemoryDB
)
