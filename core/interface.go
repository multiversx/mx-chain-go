package core

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/transaction"
)

// Core interface will abstract all the subpackage functionalities and will
//  provide access to it's members where needed
type Core interface {
	Indexer() Indexer
}

// Indexer is and interface for saving node specific data to other storages.
// This could be an elasticsearch intex, a MySql database or any other external services.
type Indexer interface {
	SaveBlock(body block.Body, header *block.Header, txPool map[string]*transaction.Transaction)
	SaveMetaBlock(metaBlock *block.MetaBlock, headerPool map[string]*block.Header)
}
