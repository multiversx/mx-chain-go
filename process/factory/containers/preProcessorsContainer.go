package containers

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/process"
)

var _ process.PreProcessorsContainer = (*preProcessorsContainer)(nil)

// preProcessorsContainer is an PreProcessors holder organized by type
type preProcessorsContainer struct {
	objects *container.MutexMap
}

// NewPreProcessorsContainer will create a new instance of a container
func NewPreProcessorsContainer() *preProcessorsContainer {
	return &preProcessorsContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (ppc *preProcessorsContainer) Get(key block.Type) (process.PreProcessor, error) {
	value, ok := ppc.objects.Get(uint8(key))
	if !ok {
		return nil, fmt.Errorf("%w in pre processor container for key %v", process.ErrInvalidContainerKey, key)
	}

	preProcessor, ok := value.(process.PreProcessor)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return preProcessor, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (ppc *preProcessorsContainer) Add(key block.Type, preProcessor process.PreProcessor) error {
	if check.IfNil(preProcessor) {
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
func (ppc *preProcessorsContainer) AddMultiple(keys []block.Type, preProcessors []process.PreProcessor) error {
	if len(keys) != len(preProcessors) {
		return process.ErrLenMismatch
	}

	for idx, key := range keys {
		err := ppc.Add(key, preProcessors[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (ppc *preProcessorsContainer) Replace(key block.Type, preProcessor process.PreProcessor) error {
	if check.IfNil(preProcessor) {
		return process.ErrNilContainerElement
	}

	ppc.objects.Set(uint8(key), preProcessor)
	return nil
}

// Remove will remove an object at a given key
func (ppc *preProcessorsContainer) Remove(key block.Type) {
	ppc.objects.Remove(uint8(key))
}

// Len returns the length of the added objects
func (ppc *preProcessorsContainer) Len() int {
	return ppc.objects.Len()
}

// Keys returns all the keys from the container
func (ppc *preProcessorsContainer) Keys() []block.Type {
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
func (ppc *preProcessorsContainer) IsInterfaceNil() bool {
	return ppc == nil
}
