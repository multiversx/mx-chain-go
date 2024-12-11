package requesterscontainer

import "github.com/multiversx/mx-chain-go/dataRetriever"

// TODO: We should implement this for meta as well and use when nodeRunner will be split in shardRunner+metaRunner

// RequesterContainerFactoryCreator should create requester container factories depending on the chain run type
type RequesterContainerFactoryCreator interface {
	CreateRequesterContainerFactory(args FactoryArgs) (dataRetriever.RequestersContainerFactory, error)
	IsInterfaceNil() bool
}
