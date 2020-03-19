package disabled

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type validityAttester struct {
}

// NewValidityAttester returns a new instance of validityAttester
func NewValidityAttester() *validityAttester {
	return &validityAttester{}
}

func (v *validityAttester) CheckBlockAgainstFinal(headerHandler data.HeaderHandler) error {
	return nil
}

func (v *validityAttester) CheckBlockAgainstRounder(headerHandler data.HeaderHandler) error {
	return nil
}

func (v *validityAttester) IsInterfaceNil() bool {
	return v == nil
}
