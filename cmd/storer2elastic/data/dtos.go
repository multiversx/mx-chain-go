package data

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/indexer"
)

// RoundPersistedData holds the data for a round which inclused the meta block and the shard headers
type RoundPersistedData struct {
	MetaBlockData *HeaderData
	ShardHeaders  map[uint32][]*HeaderData
}

// HeaderData holds the data for a shard in a round
type HeaderData struct {
	Header           data.HeaderHandler
	Body             *block.Body
	BodyTransactions *indexer.Pool
}
