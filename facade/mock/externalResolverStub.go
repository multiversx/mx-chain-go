package mock

import "github.com/ElrondNetwork/elrond-go-sandbox/node/external"

type ExternalResolverStub struct {
	RecentNotarizedBlocksCalled func(maxShardHeadersNum int) ([]external.RecentBlock, error)
}

func (ers *ExternalResolverStub) RecentNotarizedBlocks(maxShardHeadersNum int) ([]external.RecentBlock, error) {
	return ers.RecentNotarizedBlocksCalled(maxShardHeadersNum)
}
