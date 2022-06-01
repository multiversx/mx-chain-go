package blockAPI

import (
	"encoding/hex"
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
)

func (bap *baseAPIBlockProcessor) loadMiniblock(miniblockHeader data.MiniBlockHeaderHandler, epoch uint32) ([]byte, *block.MiniBlock, error) {
	miniblockHash := miniblockHeader.GetHash()
	miniblockBytes, err := bap.getFromStorerWithEpoch(dataRetriever.MiniBlockUnit, miniblockHash, epoch)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v, hash = %s", errCannotLoadMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	miniBlock := &block.MiniBlock{}
	err = bap.marshalizer.Unmarshal(miniBlock, miniblockBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v, hash = %s", errCannotUnmarshalMiniblocks, err, hex.EncodeToString(miniblockHash))
	}

	return miniblockHash, miniBlock, nil
}
