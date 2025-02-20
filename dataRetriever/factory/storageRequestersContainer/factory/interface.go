package factory

import (
	"github.com/multiversx/mx-chain-go/dataRetriever"
	storagerequesterscontainer "github.com/multiversx/mx-chain-go/dataRetriever/factory/storageRequestersContainer"
)

// ShardRequestersContainerCreatorHandler defines a creator of shard requesters container creator
type ShardRequestersContainerCreatorHandler interface {
	CreateShardRequestersContainerFactory(args storagerequesterscontainer.FactoryArgs) (dataRetriever.RequestersContainerFactory, error)
	IsInterfaceNil() bool
}
