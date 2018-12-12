package dataPool

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/shardedData"
)

type TransientDataPool struct {
	TransactionDataPool shardedData.ShardedData
}
