package headerCheck

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
)

// ComputeConsensusGroup will compute the consensus group that assembled the provided block
func ComputeConsensusGroup(header data.HeaderHandler, nodesCoordinator nodesCoordinator.NodesCoordinator) (validatorsGroup []nodesCoordinator.Validator, err error) {
	if check.IfNil(header) {
		return nil, process.ErrNilHeaderHandler
	}
	if check.IfNil(nodesCoordinator) {
		return nil, process.ErrNilNodesCoordinator
	}

	prevRandSeed := header.GetPrevRandSeed()

	// TODO: change here with an activation flag if start of epoch block needs to be validated by the new epoch nodes
	epoch := header.GetEpoch()
	if header.IsStartOfEpochBlock() && epoch > 0 {
		epoch = epoch - 1
	}

	return nodesCoordinator.ComputeConsensusGroup(prevRandSeed, header.GetRound(), header.GetShardID(), epoch)
}

// ComputeSignersPublicKeys will extract from the provided consensus group slice only the strings that matched with the bitmap
func ComputeSignersPublicKeys(consensusGroup []string, bitmap []byte) []string {
	nbBitsBitmap := len(bitmap) * 8
	consensusGroupSize := len(consensusGroup)
	size := consensusGroupSize
	if consensusGroupSize > nbBitsBitmap {
		size = nbBitsBitmap
	}

	result := make([]string, 0, len(consensusGroup))

	for i := 0; i < size; i++ {
		indexRequired := (bitmap[i/8] & (1 << uint16(i%8))) > 0
		if !indexRequired {
			continue
		}

		pubKey := consensusGroup[i]
		result = append(result, pubKey)
	}

	return result
}
