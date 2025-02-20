package interceptorscontainer

import "github.com/multiversx/mx-chain-go/process"

type shardInterceptorsContainerFactoryCreator struct {
}

// NewShardInterceptorsContainerFactoryCreator creates a shard interceptors container factory creator
func NewShardInterceptorsContainerFactoryCreator() *shardInterceptorsContainerFactoryCreator {
	return &shardInterceptorsContainerFactoryCreator{}
}

// CreateInterceptorsContainerFactory creates a shard interceptors container factory
func (f *shardInterceptorsContainerFactoryCreator) CreateInterceptorsContainerFactory(args CommonInterceptorsContainerFactoryArgs) (process.InterceptorsContainerFactory, error) {
	return NewShardInterceptorsContainerFactory(args)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *shardInterceptorsContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
