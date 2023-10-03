package disabled

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/resolvers/disabled"
)

type resolversContainer struct {
}

// NewDisabledResolversContainer returns a new instance of disabled resolvers container
func NewDisabledResolversContainer() *resolversContainer {
	return &resolversContainer{}
}

// Get returns a new instance of disabled resolver and nil
func (container *resolversContainer) Get(_ string) (dataRetriever.Resolver, error) {
	return disabled.NewDisabledResolver(), nil
}

// Add returns nil as it is disabled
func (container *resolversContainer) Add(_ string, _ dataRetriever.Resolver) error {
	return nil
}

// AddMultiple returns nil as it is disabled
func (container *resolversContainer) AddMultiple(_ []string, _ []dataRetriever.Resolver) error {
	return nil
}

// Replace returns nil as it is disabled
func (container *resolversContainer) Replace(_ string, _ dataRetriever.Resolver) error {
	return nil
}

// Remove does nothing as it is disabled
func (container *resolversContainer) Remove(_ string) {
}

// Len returns 0 as it is disabled
func (container *resolversContainer) Len() int {
	return 0
}

// ResolverKeys returns empty string as it is disabled
func (container *resolversContainer) ResolverKeys() string {
	return ""
}

// Iterate does nothing as it is disabled
func (container *resolversContainer) Iterate(_ func(key string, resolver dataRetriever.Resolver) bool) {
}

// Close returns nil as it is disabled
func (container *resolversContainer) Close() error {
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (container *resolversContainer) IsInterfaceNil() bool {
	return container == nil
}
