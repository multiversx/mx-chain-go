package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModifiedHashes_Clone(t *testing.T) {
	t.Parallel()

	mh := make(ModifiedHashes)
	mh["aaa"] = struct{}{}
	mh["bbb"] = struct{}{}

	cloned := mh.Clone()
	assert.NotEqual(t, fmt.Sprintf("%p", mh), fmt.Sprintf("%p", cloned)) //pointer testing
	assert.Equal(t, len(mh), len(cloned))

	for key := range mh {
		_, found := cloned[key]
		assert.True(t, found)
	}

}
