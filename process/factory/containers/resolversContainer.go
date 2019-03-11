package containers

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/cornelk/hashmap"
)

// ResolversContainer is a resolvers holder organized by type
type ResolversContainer struct {
	objects *hashmap.HashMap
}

// NewResolversContainer will create a new instance of a container
func NewResolversContainer() *ResolversContainer {
	return &ResolversContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (rc *ResolversContainer) Get(key string) (process.Resolver, error) {
	value, ok := rc.objects.Get(key)
	if !ok {
		return nil, process.ErrInvalidContainerKey
	}

	resolver, ok := value.(process.Resolver)
	if !ok {
		return nil, process.ErrWrongTypeInContainer
	}

	return resolver, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (rc *ResolversContainer) Add(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}

	ok := rc.objects.Insert(key, resolver)

	if !ok {
		return process.ErrContainerKeyAlreadyExists
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (rc *ResolversContainer) Replace(key string, resolver process.Resolver) error {
	if resolver == nil {
		return process.ErrNilContainerElement
	}

	rc.objects.Set(key, resolver)
	return nil
}

// Remove will remove an object at a given key
func (rc *ResolversContainer) Remove(key string) {
	rc.objects.Del(key)
}

// Len returns the length of the added objects
func (rc *ResolversContainer) Len() int {
	return rc.objects.Len()
}
