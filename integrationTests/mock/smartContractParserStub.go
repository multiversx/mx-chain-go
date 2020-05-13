package mock

import (
	"github.com/ElrondNetwork/elrond-go/genesis"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// SmartContractParserStub -
type SmartContractParserStub struct {
	InitialSmartContractsSplitOnOwnersShardsCalled func(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialSmartContractHandler, error)
	InitialSmartContractsCalled                    func() []genesis.InitialSmartContractHandler
}

// InitialSmartContractsSplitOnOwnersShards -
func (scps *SmartContractParserStub) InitialSmartContractsSplitOnOwnersShards(shardCoordinator sharding.Coordinator) (map[uint32][]genesis.InitialSmartContractHandler, error) {
	if scps.InitialSmartContractsSplitOnOwnersShardsCalled != nil {
		return scps.InitialSmartContractsSplitOnOwnersShardsCalled(shardCoordinator)
	}

	return make(map[uint32][]genesis.InitialSmartContractHandler), nil
}

// InitialSmartContracts -
func (scps *SmartContractParserStub) InitialSmartContracts() []genesis.InitialSmartContractHandler {
	if scps.InitialSmartContractsCalled != nil {
		return scps.InitialSmartContractsCalled()
	}

	return make([]genesis.InitialSmartContractHandler, 0)
}

// IsInterfaceNil -
func (scps *SmartContractParserStub) IsInterfaceNil() bool {
	return scps == nil
}
