package blockAPI

import (
	"github.com/multiversx/mx-chain-core-go/data/api"
)

func filterOutDuplicatedMiniblocks(miniblocks []*api.MiniBlock) []*api.MiniBlock {
	filteredMiniblocks := make([]*api.MiniBlock, 0, len(miniblocks))
	seenMiniblocks := make(map[string]struct{})

	for _, miniblock := range miniblocks {
		_, ok := seenMiniblocks[miniblock.Hash]
		if ok {
			continue
		}

		filteredMiniblocks = append(filteredMiniblocks, miniblock)
		seenMiniblocks[miniblock.Hash] = struct{}{}
	}

	return filteredMiniblocks
}
