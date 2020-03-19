package structs

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// ComponentsNeededForBootstrap holds the components which need to be initialized from network
type ComponentsNeededForBootstrap struct {
	EpochStartMetaBlock         *block.MetaBlock
	PreviousEpochStartMetaBlock *block.MetaBlock
	ShardHeader                 *block.Header //only for shards, nil for meta
	NodesConfig                 *sharding.NodesSetup
	ShardHeaders                map[uint32]*block.Header
	ShardCoordinator            sharding.Coordinator
	Tries                       state.TriesHolder
	PendingMiniBlocks           map[string]*block.MiniBlock
}
