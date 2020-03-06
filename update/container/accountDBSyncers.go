package containers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/update"
	"github.com/cornelk/hashmap"
)

type accountDBSyncers struct {
	objects *hashmap.HashMap
}

// NewAccountsDBSyncersContainer will create a new instance of a container
func NewAccountsDBSyncersContainer() *accountDBSyncers {
	return &accountDBSyncers{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (t *accountDBSyncers) Get(key string) (update.AccountsDBSyncer, error) {
	value, ok := t.objects.Get(key)
	if !ok {
		return nil, update.ErrInvalidContainerKey
	}

	syncer, ok := value.(update.AccountsDBSyncer)
	if !ok {
		return nil, update.ErrWrongTypeInContainer
	}

	return syncer, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (t *accountDBSyncers) Add(key string, val update.AccountsDBSyncer) error {
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
func (t *accountDBSyncers) AddMultiple(keys []string, values []update.AccountsDBSyncer) error {
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
func (t *accountDBSyncers) Replace(key string, val update.AccountsDBSyncer) error {
	if check.IfNil(val) {
		return process.ErrNilContainerElement
	}

	t.objects.Set(key, val)
	return nil
}

// Remove will remove an object at a given key
func (t *accountDBSyncers) Remove(key string) {
	t.objects.Del(key)
}

// Len returns the length of the added objects
func (t *accountDBSyncers) Len() int {
	return t.objects.Len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (t *accountDBSyncers) IsInterfaceNil() bool {
	return t == nil
}
