package address

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type specialAddresses struct {
	elrond           []byte
	ownAddress       []byte
	adrConv          state.AddressConverter
	shardCoordinator sharding.Coordinator
}

// NewSpecialAddressHolder creates a special address holder
func NewSpecialAddressHolder(
	elrond []byte,
	ownAddress []byte,
	adrConv state.AddressConverter,
	shardCoordinator sharding.Coordinator,
) (*specialAddresses, error) {
	if elrond == nil {
		return nil, data.ErrNilElrondAddress
	}
	if ownAddress == nil {
		return nil, data.ErrNilOwnAddress
	}
	if adrConv == nil {
		return nil, data.ErrNilAddressConverter
	}
	if shardCoordinator == nil {
		return nil, data.ErrNilShardCoordinator
	}

	sp := &specialAddresses{
		elrond:           elrond,
		ownAddress:       ownAddress,
		adrConv:          adrConv,
		shardCoordinator: shardCoordinator,
	}

	return sp, nil
}

// ElrondCommunityAddress provides elrond address
func (sp *specialAddresses) ElrondCommunityAddress() []byte {
	return sp.elrond
}

// OwnAddress provides own address
func (sp *specialAddresses) LeaderAddress() []byte {
	return sp.ownAddress
}

// ShardIdForAddress calculates shard id for address
func (sp *specialAddresses) ShardIdForAddress(pubKey []byte) uint32 {
	convAdr, err := sp.adrConv.CreateAddressFromPublicKeyBytes(pubKey)
	if err != nil {
		return sp.shardCoordinator.NumberOfShards()
	}

	return sp.shardCoordinator.ComputeId(convAdr)
}
