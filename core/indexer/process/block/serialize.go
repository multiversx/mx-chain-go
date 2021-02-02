package block

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

// ErrHeaderTypeAssertion signals that body type assertion failed
var ErrHeaderTypeAssertion = errors.New("elasticsearch - header type assertion failed")

// SerializeBlock will serialize block for database
func (bp *blockProcessor) SerializeBlock(elasticBlock *types.Block) (*bytes.Buffer, error) {
	blockBytes, err := json.Marshal(elasticBlock)
	if err != nil {
		return nil, err
	}

	buff := &bytes.Buffer{}

	buff.Grow(len(blockBytes))
	_, err = buff.Write(blockBytes)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// SerializeEpochInfoData will serialize information about current epoch
func (bp *blockProcessor) SerializeEpochInfoData(header data.HeaderHandler) (*bytes.Buffer, error) {
	metablock, ok := header.(*block.MetaBlock)
	if !ok {
		return nil, ErrHeaderTypeAssertion
	}

	epochInfo := &types.EpochInfo{
		AccumulatedFees: metablock.AccumulatedFeesInEpoch.String(),
		DeveloperFees:   metablock.DevFeesInEpoch.String(),
	}

	epochInfoBytes, err := json.Marshal(epochInfo)
	if err != nil {
		log.Warn("blockProcessor.SerializeEpochInfoData cannot serialize epoch info", "error", err)
		return nil, err
	}

	buff := &bytes.Buffer{}
	buff.Grow(len(epochInfoBytes))
	_, err = buff.Write(epochInfoBytes)
	if err != nil {
		log.Warn("blockProcessor.SerializeEpochInfoData cannot write in buffer", "error", err)
		return nil, err
	}

	return buff, nil
}
