package interceptorscontainer

import "github.com/multiversx/mx-chain-go/process"

type sovereignShardInterceptorsContainerFactoryCreator struct {
}

// NewSovereignShardInterceptorsContainerFactoryCreator creates a sovereign shard interceptors container factory creator
func NewSovereignShardInterceptorsContainerFactoryCreator() *sovereignShardInterceptorsContainerFactoryCreator {
	return &sovereignShardInterceptorsContainerFactoryCreator{}
}

// CreateInterceptorsContainerFactory creates a sovereign shard interceptors container factory
func (f *sovereignShardInterceptorsContainerFactoryCreator) CreateInterceptorsContainerFactory(args CommonInterceptorsContainerFactoryArgs) (process.InterceptorsContainerFactory, error) {
	shardContainer, err := NewShardInterceptorsContainerFactory(args)
	if err != nil {
		return nil, err
	}

	argsSov := ArgsSovereignShardInterceptorsContainerFactory{
		ShardContainer:           shardContainer,
		IncomingHeaderSubscriber: args.IncomingHeaderSubscriber,
	}
	return NewSovereignShardInterceptorsContainerFactory(argsSov)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (f *sovereignShardInterceptorsContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
