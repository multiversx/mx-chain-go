package groupSelectors

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/validators"
)

func (ihgs *indexHashedGroupSelector) EligibleList() []validators.Validator {
	return ihgs.eligibleList
}
