package containers

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.TrieSyncContainer = (*trieSyncers)(nil)

type trieSyncers struct {
	objects *container.MutexMap
}

// NewTrieSyncersContainer will create a new instance of a container
func NewTrieSyncersContainer() *trieSyncers {
	return &trieSyncers{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (t *trieSyncers) Get(key string) (update.TrieSyncer, error) {
	value, ok := t.objects.Get(key)
	if !ok {
		return nil, update.ErrInvalidContainerKey
	}

	syncer, ok := value.(update.TrieSyncer)
	if !ok {
		return nil, update.ErrWrongTypeInContainer
	}

	return syncer, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (t *trieSyncers) Add(key string, val update.TrieSyncer) error {
	if check.IfNil(val) {
		return update.ErrNilContainerElement
	}

	ok := t.objects.Insert(key, val)
	if !ok {
		return update.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (t *trieSyncers) AddMultiple(keys []string, values []update.TrieSyncer) error {
	if len(keys) != len(values) {
		return update.ErrLenMismatch
	}

	for idx, key := range keys {
		err := t.Add(key, values[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (t *trieSyncers) Replace(key string, val update.TrieSyncer) error {
	if check.IfNil(val) {
		return process.ErrNilContainerElement
	}

	t.objects.Set(key, val)
	return nil
}

// Remove will remove an object at a given key
func (t *trieSyncers) Remove(key string) {
	t.objects.Remove(key)
}

// Len returns the length of the added objects
func (t *trieSyncers) Len() int {
	return t.objects.Len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *trieSyncers) IsInterfaceNil() bool {
	return t == nil
}
