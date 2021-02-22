package block

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/integrationTests"
)

// StepDelay -
var StepDelay = time.Second / 10

// GetBlockProposersIndexes -
func GetBlockProposersIndexes(
	consensusMap map[uint32][]*integrationTests.TestProcessorNode,
	nodesMap map[uint32][]*integrationTests.TestProcessorNode,
) map[uint32]int {

	indexProposer := make(map[uint32]int)

	for sh, testNodeList := range nodesMap {
		for k, testNode := range testNodeList {
			if consensusMap[sh][0] == testNode {
				indexProposer[sh] = k
			}
		}
	}

	return indexProposer
}
