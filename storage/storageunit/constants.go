package storageunit

import "github.com/ElrondNetwork/elrond-go-storage/storageUnit"

const (
	LRUCache     = storageUnit.LRUCache
	SizeLRUCache = storageUnit.SizeLRUCache
)

// LvlDB currently the only supported DBs
// More to be added
const (
	LvlDB       = storageUnit.LvlDB
	LvlDBSerial = storageUnit.LvlDBSerial
	MemoryDB    = storageUnit.MemoryDB
)
