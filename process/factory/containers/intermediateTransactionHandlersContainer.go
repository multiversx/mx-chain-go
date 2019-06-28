package containers

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/cornelk/hashmap"
)

// IntermediateTransactionHandlersContainer is an IntermediateTransactionHandlers holder organized by type
type IntermediateTransactionHandlersContainer struct {
	objects *hashmap.HashMap
}

// NewIntermediateTransactionHandlersContainer will create a new instance of a container
func NewIntermediateTransactionHandlersContainer() *IntermediateTransactionHandlersContainer {
	return &IntermediateTransactionHandlersContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ppc *IntermediateTransactionHandlersContainer) Get(key block.Type) (process.IntermediateTransactionHandler, error) {
	value, ok := ppc.objects.Get(uint8(key))
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	interProcessor, ok := value.(process.IntermediateTransactionHandler)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return interProcessor, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ppc *IntermediateTransactionHandlersContainer) Add(key block.Type, interProcessor process.IntermediateTransactionHandler) error {
	if interProcessor == nil {
		return process.ErrNilContainerElement
	}

	ok := ppc.objects.Insert(uint8(key), interProcessor)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (ppc *IntermediateTransactionHandlersContainer) AddMultiple(keys []block.Type, IntermediateTransactionHandlers []process.IntermediateTransactionHandler) error {
	if len(keys) != len(IntermediateTransactionHandlers) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := ppc.Add(key, IntermediateTransactionHandlers[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ppc *IntermediateTransactionHandlersContainer) Replace(key block.Type, interProcessor process.IntermediateTransactionHandler) error {
	if interProcessor == nil {
		return process.ErrNilContainerElement
	}

	ppc.objects.Set(uint8(key), interProcessor)
	return nil
}

// Remove will remove an object at a given key
func (ppc *IntermediateTransactionHandlersContainer) Remove(key block.Type) {
	ppc.objects.Del(uint8(key))
}

// Len returns the length of the added objects
func (ppc *IntermediateTransactionHandlersContainer) Len() int {
	return ppc.objects.Len()
}

func (ppc *IntermediateTransactionHandlersContainer) Keys() []block.Type {
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
