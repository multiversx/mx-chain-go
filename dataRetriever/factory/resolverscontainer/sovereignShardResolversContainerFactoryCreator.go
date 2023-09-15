package resolverscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

type sovereignShardResolversContainerFactoryCreator struct {
}

func NewSovereignShardResolversContainerFactoryCreator() *sovereignShardResolversContainerFactoryCreator {
	return &sovereignShardResolversContainerFactoryCreator{}
}

func (f *sovereignShardResolversContainerFactoryCreator) CreateShardResolversContainerFactory(args FactoryArgs,
) (dataRetriever.ResolversContainerFactory, error) {
	shardContainer, err := NewShardResolversContainerFactory(args)
	if err != nil {
		return nil, err
	}

	return NewSovereignShardResolversContainerFactory(shardContainer)
}

func (f *sovereignShardResolversContainerFactoryCreator) IsInterfaceNil() bool {
	return f == nil
}
