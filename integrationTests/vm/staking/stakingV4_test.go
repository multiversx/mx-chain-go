package staking

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestMetaProcessor(t *testing.T) {
	node := NewTestMetaProcessor(1, 1, 1, 1, 1)
	header, err := node.MetaBlockProcessor.CreateNewHeader(0, 0)
	require.Nil(t, err)
	fmt.Println(header)
}
