package interceptorscontainer

import "github.com/multiversx/mx-chain-go/process"

// TODO: Implement this in MX-14517

// sovereignShardInterceptorsContainerFactory will handle the creation of sovereign interceptors container
type sovereignShardInterceptorsContainerFactory struct {
	*baseInterceptorsContainerFactory
}

// NewSovereignShardInterceptorsContainerFactory creates a new sovereign interceptors factory
func NewSovereignShardInterceptorsContainerFactory(
	_ CommonInterceptorsContainerFactoryArgs,
) (*sovereignShardInterceptorsContainerFactory, error) {
	return &sovereignShardInterceptorsContainerFactory{}, nil
}

// Create returns an interceptor container that will hold all sovereign interceptors
func (sicf *sovereignShardInterceptorsContainerFactory) Create() (process.InterceptorsContainer, process.InterceptorsContainer, error) {
	return nil, nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sicf *sovereignShardInterceptorsContainerFactory) IsInterfaceNil() bool {
	return sicf == nil
}
