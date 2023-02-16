package containers

import (
	"fmt"
	"sort"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/container"
	"github.com/multiversx/mx-chain-go/dataRetriever"
)

var _ dataRetriever.RequestersContainer = (*requestersContainer)(nil)

// requestersContainer is a requesters holder organized by type
// TODO: extract resolversContainer + requestersContainer in a general container struct
type requestersContainer struct {
	objects *container.MutexMap
}

// NewRequestersContainer will create a new instance of a container
func NewRequestersContainer() *requestersContainer {
	return &requestersContainer{
		objects: container.NewMutexMap(),
	}
}

// Get returns the object stored at a certain key.
// Returns an error if the element does not exist
func (rc *requestersContainer) Get(key string) (dataRetriever.Requester, error) {
	value, ok := rc.objects.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w in requesters container for key %v", dataRetriever.ErrInvalidContainerKey, key)
	}

	requester, ok := value.(dataRetriever.Requester)
	if !ok {
		return nil, dataRetriever.ErrWrongTypeInContainer
	}

	return requester, nil
}

// Add will add an object at a given key. Returns
// an error if the element already exists
func (rc *requestersContainer) Add(key string, requester dataRetriever.Requester) error {
	if check.IfNil(requester) {
		return dataRetriever.ErrNilContainerElement
	}

	ok := rc.objects.Insert(key, requester)
	if !ok {
		return dataRetriever.ErrContainerKeyAlreadyExists
	}

	return nil
}

// AddMultiple will add objects with given keys. Returns
// an error if one element already exists, lengths mismatch or an interceptor is nil
func (rc *requestersContainer) AddMultiple(keys []string, requesters []dataRetriever.Requester) error {
	if len(keys) != len(requesters) {
		return dataRetriever.ErrLenMismatch
	}

	for idx, key := range keys {
		err := rc.Add(key, requesters[idx])
		if err != nil {
			return err
		}
	}

	return nil
}

// Replace will add (or replace if it already exists) an object at a given key
func (rc *requestersContainer) Replace(key string, requester dataRetriever.Requester) error {
	if check.IfNil(requester) {
		return dataRetriever.ErrNilContainerElement
	}

	rc.objects.Set(key, requester)
	return nil
}

// Remove will remove an object at a given key
func (rc *requestersContainer) Remove(key string) {
	rc.objects.Remove(key)
}

// Len returns the length of the added objects
func (rc *requestersContainer) Len() int {
	return rc.objects.Len()
}

// RequesterKeys will return the contained requesters' keys in a concatenated string
func (rc *requestersContainer) RequesterKeys() string {
	keys := rc.objects.Keys()

	stringKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		key, ok := k.(string)
		if !ok {
			continue
		}

		stringKeys = append(stringKeys, key)
	}

	sort.Slice(stringKeys, func(i, j int) bool {
		return strings.Compare(stringKeys[i], stringKeys[j]) < 0
	})

	return strings.Join(stringKeys, ", ")
}

// Iterate will call the provided handler for each and every key-value pair
func (rc *requestersContainer) Iterate(handler func(key string, requester dataRetriever.Requester) bool) {
	if handler == nil {
		return
	}

	for _, keyVal := range rc.objects.Keys() {
		key, ok := keyVal.(string)
		if !ok {
			continue
		}

		val, ok := rc.objects.Get(key)
		if !ok {
			continue
		}

		requester, ok := val.(dataRetriever.Requester)
		if !ok {
			continue
		}

		shouldContinue := handler(key, requester)
		if !shouldContinue {
			return
		}
	}
}

// Close will return nil
func (rc *requestersContainer) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (rc *requestersContainer) IsInterfaceNil() bool {
	return rc == nil
}
