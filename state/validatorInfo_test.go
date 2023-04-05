package state

import (
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/stretchr/testify/assert"
)

func TestValidatorInfo_IsInterfaceNile(t *testing.T) {
	t.Parallel()

	vi := &ValidatorInfo{}
	assert.False(t, check.IfNil(vi))
}

func TestTestTes(t *testing.T) {
	type abc struct {
		ff int
	}

	obj := abc{ff: 37}
	objCopy := *&obj
	fmt.Printf("%p\n", &obj)
	fmt.Printf("%p\n", &objCopy)
}
