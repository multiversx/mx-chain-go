package blockPool

import (
	"encoding/binary"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
	"sync"
)

var defaultCacherConfig = &storage.CacheConfig{
	Size: 100,
	Type: storage.LRUCache,
}

// BlockPool holds cashers of headers and bodies
type BlockPool struct {
	HeaderStore storage.Cacher
	BodyStore   storage.Cacher

	cacherConfig *storage.CacheConfig

	headerHandlerLock sync.RWMutex
	addHeaderHandlers []func(nounce uint64)

	bodyHandlerLock sync.RWMutex
	addBodyHandlers []func(nounce uint64)
}

// NewBlockPool creates a BlockPool object which contains an empty pool of headers and bodies
func NewBlockPool(cacherConfig *storage.CacheConfig) *BlockPool {
	if cacherConfig == nil {
		cacherConfig = defaultCacherConfig
	}

	bp := &BlockPool{cacherConfig: cacherConfig}

	var err error

	bp.HeaderStore, err = storage.NewCache(cacherConfig.Type, cacherConfig.Size)

	if err != nil {
		return nil
	}

	bp.BodyStore, err = storage.NewCache(cacherConfig.Type, cacherConfig.Size)

	if err != nil {
		return nil
	}

	return bp
}

// MiniPoolHeaderStore method returns the minipool header store which contains block nounces and headers
func (bp *BlockPool) MiniPoolHeaderStore() (c storage.Cacher) {
	return bp.HeaderStore
}

// MiniPoolBodyStore method returns the minipool body store which contains block nounces and bodies
func (bp *BlockPool) MiniPoolBodyStore() (c storage.Cacher) {
	return bp.BodyStore
}

// AddHeader method adds a block header to the corresponding pool
func (bp *BlockPool) AddHeader(nounce uint64, hdr *block.Header) {
	mp := bp.MiniPoolHeaderStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	found, _ := mp.HasOrAdd(key, hdr)

	if bp.addHeaderHandlers != nil && !found {
		for _, handler := range bp.addHeaderHandlers {
			go handler(nounce)
		}
	}
}

// AddBody method adds a block body to the corresponding pool
func (bp *BlockPool) AddBody(nounce uint64, blk *block.Block) {
	mp := bp.MiniPoolBodyStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	found, _ := mp.HasOrAdd(key, blk)

	if bp.addBodyHandlers != nil && !found {
		for _, handler := range bp.addBodyHandlers {
			go handler(nounce)
		}
	}
}

// Header method returns the block header with given nounce
func (bp *BlockPool) Header(nounce uint64) (interface{}, bool) {
	mp := bp.MiniPoolHeaderStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	return mp.Get(key)
}

// Body method returns the block body with given nounce
func (bp *BlockPool) Body(nounce uint64) (interface{}, bool) {
	mp := bp.MiniPoolBodyStore()

	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	return mp.Get(key)
}

// RemoveHeader method removes a block header with given nounce
func (bp *BlockPool) RemoveHeader(nounce uint64) {
	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	bp.HeaderStore.Remove(key)
}

// RemoveBody method removes a block body with given nounce
func (bp *BlockPool) RemoveBody(nounce uint64) {
	key := make([]byte, 8)
	binary.PutUvarint(key, nounce)

	bp.BodyStore.Remove(key)
}

// ClearHeaderMiniPool deletes all the block headers from the coresponding pool
func (bp *BlockPool) ClearHeaderMiniPool() {
	bp.HeaderStore.Clear()
}

// ClearBodyMiniPool deletes all the block bodies from the coresponding pool
func (bp *BlockPool) ClearBodyMiniPool() {
	bp.BodyStore.Clear()
}

// RegisterHeaderHandler will register a new call back function wich will be notified when a new header will be added
func (bp *BlockPool) RegisterHeaderHandler(headerHandler func(nounce uint64)) {
	bp.headerHandlerLock.Lock()

	if bp.addHeaderHandlers == nil {
		bp.addHeaderHandlers = make([]func(nounce uint64), 0)
	}

	bp.addHeaderHandlers = append(bp.addHeaderHandlers, headerHandler)

	bp.headerHandlerLock.Unlock()
}

// RegisterBodyHandler will register a new call back function wich will be notified when a new body will be added
func (bp *BlockPool) RegisterBodyHandler(bodyHandler func(nounce uint64)) {
	bp.bodyHandlerLock.Lock()

	if bp.addBodyHandlers == nil {
		bp.addBodyHandlers = make([]func(nounce uint64), 0)
	}

	bp.addBodyHandlers = append(bp.addBodyHandlers, bodyHandler)

	bp.bodyHandlerLock.Unlock()
}
