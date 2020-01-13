package containers

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/cornelk/hashmap"
)

type accounts struct {
	objects *hashmap.HashMap
}

// NewAccountsHandlerContainer will create a new instance of a container
func NewAccountsHandlerContainer() *accounts {
	return &accounts{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (a *accounts) Get(key string) (state.AccountsAdapter, error) {
	value, ok := a.objects.Get(key)
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	accountsAdapter, ok := value.(state.AccountsAdapter)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return accountsAdapter, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (a *accounts) Add(key string, val state.AccountsAdapter) error {
	if check.IfNil(val) {
		return process.ErrNilContainerElement
	}

	ok := a.objects.Insert(key, val)
	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (a *accounts) AddMultiple(keys []string, values []state.AccountsAdapter) error {
	if len(keys) != len(values) {
		return process.ErrLenMismatch
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
func (a *accounts) Replace(key string, val state.AccountsAdapter) error {
	if check.IfNil(val) {
		return process.ErrNilContainerElement
	}

	a.objects.Set(key, val)
	return nil
}

// Remove will remove an object at a given key
func (a *accounts) Remove(key string) {
	a.objects.Del(key)
}

// Len returns the length of the added objects
func (a *accounts) Len() int {
	return a.objects.Len()
}

// IsInterfaceNil returns true if there is no value under the interface
func (a *accounts) IsInterfaceNil() bool {
	return a == nil
}
