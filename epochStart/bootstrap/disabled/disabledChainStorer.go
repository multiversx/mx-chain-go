package disabled

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var _ dataRetriever.StorageService = (*chainStorer)(nil)

// ChainStorer is a mock implementation of the ChainStorer interface
type chainStorer struct {
	mapStorages map[dataRetriever.UnitType]storage.Storer
	mutex       sync.Mutex
}

// NewChainStorer -
func NewChainStorer() *chainStorer {
	return &chainStorer{
		mapStorages: make(map[dataRetriever.UnitType]storage.Storer),
	}
}

// CloseAll -
func (c *chainStorer) CloseAll() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, store := range c.mapStorages {
		err := store.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// AddStorer will add a new storer to the chain map
func (c *chainStorer) AddStorer(key dataRetriever.UnitType, s storage.Storer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.mapStorages[key] = s
}

// GetStorer returns the storer from the chain map or nil if the storer was not found
func (c *chainStorer) GetStorer(unitType dataRetriever.UnitType) storage.Storer {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.mapStorages[unitType]
	if !ok {
		c.mapStorages[unitType] = CreateMemUnit()
	}

	store := c.mapStorages[unitType]
	return store
}

// SetEpochForPutOperation won't do anything
func (c *chainStorer) SetEpochForPutOperation(epoch uint32) {
}

// Has returns true if the key is found in the selected Unit or false otherwise
// It can return an error if the provided unit type is not supported or if the
// underlying implementation of the storage unit reports an error.
func (c *chainStorer) Has(unitType dataRetriever.UnitType, key []byte) error {
	store := c.GetStorer(unitType)
	return store.Has(key)
}

// Get returns the value for the given key if found in the selected storage unit,
// nil otherwise. It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (c *chainStorer) Get(unitType dataRetriever.UnitType, key []byte) ([]byte, error) {
	store := c.GetStorer(unitType)
	return store.Get(key)
}

// Put stores the key, value pair in the selected storage unit
// It can return an error if the provided unit type is not supported
// or if the storage unit underlying implementation reports an error
func (c *chainStorer) Put(unitType dataRetriever.UnitType, key []byte, value []byte) error {
	store := c.GetStorer(unitType)
	return store.Put(key, value)
}

// GetAll gets all the elements with keys in the keys array, from the selected storage unit
// It can report an error if the provided unit type is not supported, if there is a missing
// key in the unit, or if the underlying implementation of the storage unit reports an error.
func (c *chainStorer) GetAll(unitType dataRetriever.UnitType, keys [][]byte) (map[string][]byte, error) {
	store := c.GetStorer(unitType)
	allValues := make(map[string][]byte, len(keys))

	for _, key := range keys {
		value, err := store.Get(key)
		if err != nil {
			return nil, err
		}

		allValues[string(key)] = value
	}

	return allValues, nil
}

// Destroy removes the underlying files/resources used by the storage service
func (c *chainStorer) Destroy() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, store := range c.mapStorages {
		err := store.DestroyUnit()
		if err != nil {
			return err
		}
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (c *chainStorer) IsInterfaceNil() bool {
	return c == nil
}
