package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
)

// InterceptorContainer is a struct that defines the beahaviour for a container
//  holding a list of interceptors organized by type
type InterceptorContainer struct {
	GetCalled     func(key string) (process.Interceptor, error)
	AddCalled     func(key string, interceptor process.Interceptor) error
	ReplaceCalled func(key string, interceptor process.Interceptor) error
	RemoveCalled  func(key string)
	LenCalled     func() int
}

// Get returns the interceptor stored at a certain key.
//  Returns an error if the element does not exist
func (i *InterceptorContainer) Get(key string) (process.Interceptor, error) {
	return i.GetCalled(key)
}

// Add will add an interceptor at a given key. Returns
//  an error if the element already exists
func (i *InterceptorContainer) Add(key string, interceptor process.Interceptor) error {
	return i.AddCalled(key, interceptor)
}

// Replace will add (or replace if it already exists) an interceptor at a given key
func (i *InterceptorContainer) Replace(key string, interceptor process.Interceptor) error {
	return i.ReplaceCalled(key, interceptor)
}

// Remove will remove an interceptor at a given key
func (i *InterceptorContainer) Remove(key string) {
	i.RemoveCalled(key)
}

// Len returns the length of the added interceptors
func (i *InterceptorContainer) Len() int {
	return i.LenCalled()
}
