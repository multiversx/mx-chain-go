package nodesCoordinator

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
)

func (ihgs *indexHashedNodesCoordinator) EligibleList() []consensus.Validator {
	return ihgs.nodesMap[ihgs.shardId]
}
