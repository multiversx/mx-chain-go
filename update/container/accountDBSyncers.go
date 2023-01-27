package containers

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/update"
)

var _ update.AccountsDBSyncContainer = (*accountDBSyncers)(nil)

type accountDBSyncers struct {
	objects *container.MutexMap
}

// NewAccountsDBSyncersContainer will create a new instance of a container
func NewAccountsDBSyncersContainer() *accountDBSyncers {
	return &accountDBSyncers{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (a *accountDBSyncers) Get(key string) (update.AccountsDBSyncer, error) {
	value, ok := a.objects.Get(key)
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
func (a *accountDBSyncers) Add(key string, val update.AccountsDBSyncer) error {
	if check.IfNil(val) {
		return update.ErrNilContainerElement
	}

	ok := a.objects.Insert(key, val)
	if !ok {
		return update.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (a *accountDBSyncers) AddMultiple(keys []string, values []update.AccountsDBSyncer) error {
	if len(keys) != len(values) {
		return update.ErrLenMismatch
	}

	for idx, key := range keys {
		err := a.Add(key, values[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (a *accountDBSyncers) Replace(key string, val update.AccountsDBSyncer) error {
	if check.IfNil(val) {
		return process.ErrNilContainerElement
	}

	a.objects.Set(key, val)
	return nil
}

// Remove will remove an object at a given key
func (a *accountDBSyncers) Remove(key string) {
	a.objects.Remove(key)
}

// Len returns the length of the added objects
func (a *accountDBSyncers) Len() int {
	return a.objects.Len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *accountDBSyncers) IsInterfaceNil() bool {
	return a == nil
}
