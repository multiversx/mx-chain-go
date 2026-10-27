package leavesRetriever

import (
	"context"
	"testing"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/stretchr/testify/require"
)

func TestDisabledRetrieverReportsUnavailable(t *testing.T) {
	leaves, next, err := NewDisabledLeavesRetriever().GetLeaves(1, nil, nil, context.Background())
	require.ErrorIs(t, err, common.ErrTrieLeavesRetrieverDisabled)
	require.Nil(t, leaves)
	require.Nil(t, next)
}
