package peer

import (
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/storage"
)

// DataPool indicates the main funcitonality needed in order to fetch the required blocks from the pool
type DataPool interface {
	MetaBlocks() storage.Cacher
	IsInterfaceNil() bool
}

// shardMetaMediator implementations will act as proxies whenever a decision has to be made of
//  executing a logic dependent on the chain we are currently in
type shardMetaMediator interface {
	loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error
}

// shardMetaMediated is an interface describing the internal API that needs to be provided in order
//  for shardMetaMediator implementations to be able to proxy towards the right handlers
type shardMetaMediated interface {
	loadPreviousShardHeaders(header, previousHeader *block.MetaBlock) error
	loadPreviousShardHeadersMeta(header *block.MetaBlock) error


}
