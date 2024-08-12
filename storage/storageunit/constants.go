package storageunit

import (
	"github.com/multiversx/mx-chain-storage-go/common"
)

const (
	// LRUCache defines a cache identifier with least-recently-used eviction mechanism
	LRUCache = common.LRUCache
	// SizeLRUCache defines a cache identifier with least-recently-used eviction mechanism and fixed size in bytes
	SizeLRUCache = common.SizeLRUCache
)

// DB types that are currently supported
const (
	// LvlDB represents a levelDB storage identifier
	LvlDB = common.LvlDB
	// LvlDBSerial represents a levelDB storage with serialized operations identifier
	LvlDBSerial = common.LvlDBSerial
	// MemoryDB represents an in memory storage identifier
	MemoryDB = common.MemoryDB
)

// Shard id provider types that are currently supported
const (
	BinarySplit = common.BinarySplit
)
