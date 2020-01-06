package mock

import "fmt"

type PathManagerStub struct {
	PathForEpochCalled  func(shardId string, epoch uint32, identifier string) string
	PathForStaticCalled func(shardId string, identifier string) string
}

func (p *PathManagerStub) PathForEpoch(shardId string, epoch uint32, identifier string) string {
	if p.PathForEpochCalled != nil {
		return p.PathForEpochCalled(shardId, epoch, identifier)
	}

	return fmt.Sprintf("Epoch_%d/Shard_%s/%s", epoch, shardId, identifier)
}

func (p *PathManagerStub) PathForStatic(shardId string, identifier string) string {
	if p.PathForEpochCalled != nil {
		return p.PathForStaticCalled(shardId, identifier)
	}

	return fmt.Sprintf("Static/Shard_%s/%s", shardId, identifier)
}

func (p *PathManagerStub) IsInterfaceNil() bool {
	return p == nil
}
