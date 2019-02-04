package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// ResolverContainer is a struct that defines the beahaviour for a container
//  holding a list of resolvers organized by type
type ResolverContainer struct {
	GetCalled     func(key string) (process.Resolver, error)
	AddCalled     func(key string, resolver process.Resolver) error
	ReplaceCalled func(key string, resolver process.Resolver) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

// Get returns the resolver stored at a certain key.
//  Returns an error if the element does not exist
func (i *ResolverContainer) Get(key string) (process.Resolver, error) {
	return i.GetCalled(key)
}

// Add will add a resolver at a given key. Returns
//  an error if the element already exists
func (i *ResolverContainer) Add(key string, resolver process.Resolver) error {
	return i.AddCalled(key, resolver)
}

// Replace will add (or replace if it already exists) a resolver at a given key
func (i *ResolverContainer) Replace(key string, resolver process.Resolver) error {
	return i.ReplaceCalled(key, resolver)
}

// Remove will remove a resolver at a given key
func (i *ResolverContainer) Remove(key string) {
	i.RemoveCalled(key)
}

// Len returns the length of the added resolvers
func (i *ResolverContainer) Len() int {
	return i.LenCalled()
}
