package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

type InterceptedDataStub struct {
	CheckValidCalled              func() error
	IsAddressedToOtherShardCalled func(shardCoordinator sharding.Coordinator) bool
}

func (ids InterceptedDataStub) CheckValid() error {
	return ids.CheckValidCalled()
}

func (ids InterceptedDataStub) IsAddressedToOtherShard(shardCoordinator sharding.Coordinator) bool {
	return ids.IsAddressedToOtherShardCalled(shardCoordinator)
}
