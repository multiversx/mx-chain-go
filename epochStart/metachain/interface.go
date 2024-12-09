package metachain

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/display"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/state"
)

// AuctionListDisplayHandler should be able to display auction list data during selection process
type AuctionListDisplayHandler interface {
	DisplayOwnersData(ownersData map[string]*OwnerAuctionData)
	DisplayOwnersSelectedNodes(ownersData map[string]*OwnerAuctionData)
	DisplayAuctionList(
		auctionList []state.ValidatorInfoHandler,
		ownersData map[string]*OwnerAuctionData,
		numOfSelectedNodes uint32,
	)
	IsInterfaceNil() bool
}

// TableDisplayHandler should be able to display tables in log
type TableDisplayHandler interface {
	DisplayTable(tableHeader []string, lines []*display.LineData, message string)
	IsInterfaceNil() bool
}

// ExtendedShardCoordinatorHandler defines an extended version over shard coordinator
type ExtendedShardCoordinatorHandler interface {
	sharding.Coordinator
	TotalNumberOfShards() uint32
}

type baseEconomicsHandler interface {
	startNoncePerShardFromEpochStart(epoch uint32) (map[uint32]uint64, data.MetaHeaderHandler, error)
	startNoncePerShardFromLastCrossNotarized(metaNonce uint64, epochStart data.EpochStartHandler) (map[uint32]uint64, error)
	computeNumOfTotalCreatedBlocks(
		mapStartNonce map[uint32]uint64,
		mapEndNonce map[uint32]uint64,
	) uint64
	maxPossibleNotarizedBlocks(currentRound uint64, prev data.MetaHeaderHandler) uint64
}
