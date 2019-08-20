package containers

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/cornelk/hashmap"
)

// resolversContainer is a resolvers holder organized by type
type resolversContainer struct {
	objects *hashmap.HashMap
}

// NewResolversContainer will create a new instance of a container
func NewResolversContainer() *resolversContainer {
	return &resolversContainer{
		objects: &hashmap.HashMap{},
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (rc *resolversContainer) Get(key string) (dataRetriever.Resolver, error) {
	value, ok := rc.objects.Get(key)
	if !ok {
		return nil, dataRetriever.ErrInvalidContainerKey
	}

	resolver, ok := value.(dataRetriever.Resolver)
	if !ok {
		return nil, dataRetriever.ErrWrongTypeInContainer
	}

	return resolver, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (rc *resolversContainer) Add(key string, resolver dataRetriever.Resolver) error {
	if resolver == nil {
		return dataRetriever.ErrNilContainerElement
	}

	ok := rc.objects.Insert(key, resolver)

	if !ok {
		return dataRetriever.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (rc *resolversContainer) AddMultiple(keys []string, resolvers []dataRetriever.Resolver) error {
	if len(keys) != len(resolvers) {
		return dataRetriever.ErrLenMismatch
	}

	for idx, key := range keys {
		err := rc.Add(key, resolvers[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (rc *resolversContainer) Replace(key string, resolver dataRetriever.Resolver) error {
	if resolver == nil {
		return dataRetriever.ErrNilContainerElement
	}

	rc.objects.Set(key, resolver)
	return nil
}

// Remove will remove an object at a given key
func (rc *resolversContainer) Remove(key string) {
	rc.objects.Del(key)
}

// Len returns the length of the added objects
func (rc *resolversContainer) Len() int {
	return rc.objects.Len()
}
