package containers

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/core/container"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/process"
)

var _ process.IntermediateProcessorContainer = (*intermediateTransactionHandlersContainer)(nil)

// intermediateTransactionHandlersContainer is an IntermediateTransactionHandlers holder organized by type
type intermediateTransactionHandlersContainer struct {
	objects *container.MutexMap
}

// NewIntermediateTransactionHandlersContainer will create a new instance of a container
func NewIntermediateTransactionHandlersContainer() *intermediateTransactionHandlersContainer {
	return &intermediateTransactionHandlersContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ppc *intermediateTransactionHandlersContainer) Get(key block.Type) (process.IntermediateTransactionHandler, error) {
	value, ok := ppc.objects.Get(uint8(key))
	if !ok {
		return nil, fmt.Errorf("%w in intermediate transaction handlers container for key %v", process.ErrInvalidContainerKey, key)
	}

	interProcessor, ok := value.(process.IntermediateTransactionHandler)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return interProcessor, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ppc *intermediateTransactionHandlersContainer) Add(key block.Type, interProcessor process.IntermediateTransactionHandler) error {
	if check.IfNil(interProcessor) {
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
func (ppc *intermediateTransactionHandlersContainer) AddMultiple(keys []block.Type, intermediateTransactionHandlers []process.IntermediateTransactionHandler) error {
	if len(keys) != len(intermediateTransactionHandlers) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := ppc.Add(key, intermediateTransactionHandlers[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ppc *intermediateTransactionHandlersContainer) Replace(key block.Type, interProcessor process.IntermediateTransactionHandler) error {
	if check.IfNil(interProcessor) {
		return process.ErrNilContainerElement
	}

	ppc.objects.Set(uint8(key), interProcessor)
	return nil
}

// Remove will remove an object at a given key
func (ppc *intermediateTransactionHandlersContainer) Remove(key block.Type) {
	ppc.objects.Remove(uint8(key))
}

// Len returns the length of the added objects
func (ppc *intermediateTransactionHandlersContainer) Len() int {
	return ppc.objects.Len()
}

// Keys returns all the existing keys in the container
func (ppc *intermediateTransactionHandlersContainer) Keys() []block.Type {
	keys := ppc.objects.Keys()
	keysByte := make([]block.Type, 0, len(keys))
	for _, k := range keys {
		key, ok := k.(byte)
		if !ok {
			continue
		}
		keysByte = append(keysByte, block.Type(key))
	}

	return keysByte
}

// IsInterfaceNil returns true if there is no value under the interface
func (ppc *intermediateTransactionHandlersContainer) IsInterfaceNil() bool {
	return ppc == nil
}
