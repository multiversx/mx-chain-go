package mock

import "github.com/ElrondNetwork/elrond-go/sharding"

// EligibleListProviderStub -
type EligibleListProviderStub struct {
	GetNodesPerShardCalled func(epoch uint32) (map[uint32][]sharding.Validator, error)
}

// GetNodesPerShard -
func (elps *EligibleListProviderStub) GetNodesPerShard(epoch uint32) (map[uint32][]sharding.Validator, error) {
	if elps.GetNodesPerShardCalled != nil {
		return elps.GetNodesPerShardCalled(epoch)
	}

	return nil, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (elps *EligibleListProviderStub) IsInterfaceNil() bool {
	return elps == nil
}
