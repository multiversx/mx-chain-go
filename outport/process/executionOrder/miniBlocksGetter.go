package executionOrder

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/storage"
)

type miniBlocksGetter struct {
	storer     storage.Storer
	marshaller marshal.Marshalizer
}

func newMiniblocksGetter(storer storage.Storer, marshaller marshal.Marshalizer) *miniBlocksGetter {
	return &miniBlocksGetter{
		storer:     storer,
		marshaller: marshaller,
	}
}

// GetScheduledMBs will return the scheduled miniblocks
func (bg *miniBlocksGetter) GetScheduledMBs(currentHeader, prevHeader data.HeaderHandler) ([]*block.MiniBlock, error) {
	scheduledMbs := make([]*block.MiniBlock, 0)
	if !shouldProcessPrevHeaderMiniblocks(currentHeader) {
		return scheduledMbs, nil
	}

	for _, mbHeader := range prevHeader.GetMiniBlockHeaderHandlers() {
		isScheduled := mbHeader.GetProcessingType() == int32(block.Scheduled)
		if !isScheduled {
			continue
		}

		mb, errGet := bg.getMBByHash(mbHeader.GetHash())
		if errGet != nil {
			return nil, errGet
		}

		scheduledMbs = append(scheduledMbs, mb)
	}

	return scheduledMbs, nil
}

func shouldProcessPrevHeaderMiniblocks(currentHeader data.HeaderHandler) bool {
	for _, mb := range currentHeader.GetMiniBlockHeaderHandlers() {
		mbType := block.Type(mb.GetTypeInt32())
		if mbType == block.InvalidBlock || mbType == block.SmartContractResultBlock {
			return true
		}
	}
	return false
}

func (bg *miniBlocksGetter) getMBByHash(mbHash []byte) (*block.MiniBlock, error) {
	mbBytes, err := bg.storer.Get(mbHash)
	if err != nil {
		return nil, err
	}

	mb := &block.MiniBlock{}
	err = bg.marshaller.Unmarshal(mb, mbBytes)
	if err != nil {
		return nil, err
	}

	return mb, nil
}
