package state

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestValidatorInfo_IsInterfaceNile(t *testing.T) {
	t.Parallel()

	vi := &ValidatorInfo{}
	assert.False(t, check.IfNil(vi))
}
