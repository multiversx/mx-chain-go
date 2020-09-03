package indexer

import (
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

func getBlockParamsAndBody(item *workItem) (*saveBlockData, *block.Body, bool) {
	saveBlockParams, ok := item.Data.(*saveBlockData)
	if !ok {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrInvalidWorkItemData.Error())
		return nil, nil, false
	}

	if len(saveBlockParams.signersIndexes) == 0 {
		log.Warn("block has no signers, returning")
		return nil, nil, false
	}

	body := saveBlockParams.bodyHandler.(*block.Body)
	if !ok {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrBodyTypeAssertion.Error())
		return nil, nil, false
	}

	if check.IfNil(saveBlockParams.headerHandler) {
		log.Warn("dataDispatcher.saveBlock", "removing item from queue", ErrNoHeader.Error())
		return nil, nil, false
	}

	return saveBlockParams, body, true
}
