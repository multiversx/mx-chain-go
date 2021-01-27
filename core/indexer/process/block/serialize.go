package block

import (
	"bytes"
	"encoding/json"

	"github.com/ElrondNetwork/elrond-go/core/indexer/types"
)

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
