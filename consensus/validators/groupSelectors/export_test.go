package groupSelectors

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
)

func (ihgs *indexHashedGroupSelector) EligibleList() []consensus.Validator {
	return ihgs.eligibleList
}
