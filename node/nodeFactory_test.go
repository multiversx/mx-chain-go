package node

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewNodeFactory(t *testing.T) {
	t.Parallel()

	nodeFactory := NewNodeFactory()
	require.False(t, nodeFactory.IsInterfaceNil())
}

func TestNodeFactory_CreateNewNode(t *testing.T) {
	t.Parallel()

	nodeFactory := NewNodeFactory()

	n, err := nodeFactory.CreateNewNode()
	assert.Nil(t, err)
	assert.NotNil(t, n)
}
