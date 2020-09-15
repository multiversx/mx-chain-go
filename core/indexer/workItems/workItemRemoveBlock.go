package workItems

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

type itemRemoveBlock struct {
	indexer       removeIndexer
	bodyHandler   data.BodyHandler
	headerHandler data.HeaderHandler
}

// NewItemRemoveBlock will create a new instance of itemRemoveBlock
func NewItemRemoveBlock(
	indexer removeIndexer,
	bodyHandler data.BodyHandler,
	headerHandler data.HeaderHandler,
) WorkItemHandler {
	return &itemRemoveBlock{
		indexer:       indexer,
		bodyHandler:   bodyHandler,
		headerHandler: headerHandler,
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (wirb *itemRemoveBlock) IsInterfaceNil() bool {
	return wirb == nil
}

// Save wil remove a block and miniblocks from elasticsearch database
func (wirb *itemRemoveBlock) Save() error {
	err := wirb.indexer.RemoveHeader(wirb.headerHandler)
	if err != nil {
		log.Warn("itemRemoveBlock.Save", "could not remove block", err.Error())
		return err
	}

	body, ok := wirb.bodyHandler.(*block.Body)
	if !ok {
		log.Warn("itemRemoveBlock.Save", "body", ErrBodyTypeAssertion.Error())
		return ErrBodyTypeAssertion
	}

	err = wirb.indexer.RemoveMiniblocks(wirb.headerHandler, body)
	if err != nil {
		log.Warn("itemRemoveBlock.Save", "could not remove miniblocks", err.Error())
		return err
	}

	return nil
}
