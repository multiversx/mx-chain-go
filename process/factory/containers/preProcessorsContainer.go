package containers

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/cornelk/hashmap"
)

// PreProcessorsContainer is an PreProcessors holder organized by type
type PreProcessorsContainer struct {
	objects *hashmap.HashMap
}

// NewPreProcessorsContainer will create a new instance of a container
func NewPreProcessorsContainer() *PreProcessorsContainer {
	return &PreProcessorsContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ppc *PreProcessorsContainer) Get(key block.Type) (process.PreProcessor, error) {
	value, ok := ppc.objects.Get(uint8(key))
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	preProcessor, ok := value.(process.PreProcessor)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return preProcessor, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ppc *PreProcessorsContainer) Add(key block.Type, preProcessor process.PreProcessor) error {
	if preProcessor == nil {
		return process.ErrNilContainerElement
	}

	ok := ppc.objects.Insert(uint8(key), preProcessor)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (ppc *PreProcessorsContainer) AddMultiple(keys []block.Type, PreProcessors []process.PreProcessor) error {
	if len(keys) != len(PreProcessors) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := ppc.Add(key, PreProcessors[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ppc *PreProcessorsContainer) Replace(key block.Type, preProcessor process.PreProcessor) error {
	if preProcessor == nil {
		return process.ErrNilContainerElement
	}

	ppc.objects.Set(uint8(key), preProcessor)
	return nil
}

// Remove will remove an object at a given key
func (ppc *PreProcessorsContainer) Remove(key block.Type) {
	ppc.objects.Del(uint8(key))
}

// Len returns the length of the added objects
func (ppc *PreProcessorsContainer) Len() int {
	return ppc.objects.Len()
}

func (ppc *PreProcessorsContainer) Keys() []block.Type {
	keys := make([]block.Type, 0)
	for key := range ppc.objects.Iter() {
		uint8key, ok := key.Key.(uint8)
		if !ok {
			continue
		}

		blockType := block.Type(uint8key)
		keys = append(keys, blockType)
	}
	return keys
}
