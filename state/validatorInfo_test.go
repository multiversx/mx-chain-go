package state

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestValidatorInfo_IsInterfaceNile(t *testing.T) {
	t.Parallel()

	vi := &ValidatorInfo{}
	assert.False(t, check.IfNil(vi))
}
