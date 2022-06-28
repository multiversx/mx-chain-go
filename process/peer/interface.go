package peer

import (
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/state"
)

// DataPool indicates the main functionality needed in order to fetch the required blocks from the pool
type DataPool interface {
	Headers() dataRetriever.HeadersPool
	IsInterfaceNil() bool
}

// StakingDataProviderAPI is able to provide staking data from the system smart contracts
type StakingDataProviderAPI interface {
	ComputeUnQualifiedNodes(validatorInfos state.ShardValidatorsInfoMapHandler) ([][]byte, map[string][][]byte, error)
	FillValidatorInfo(validator state.ValidatorInfoHandler) error
	GetOwnersData() map[string]*epochStart.OwnerData
	Clean()
	IsInterfaceNil() bool
}
