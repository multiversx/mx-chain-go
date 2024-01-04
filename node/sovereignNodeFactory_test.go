package node

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewSovereignNodeFactory(t *testing.T) {
	t.Parallel()

	sovereignNodeFactory := NewSovereignNodeFactory()
	require.False(t, sovereignNodeFactory.IsInterfaceNil())
}

func TestSovereignNodeFactory_CreateNewNode(t *testing.T) {
	t.Parallel()

	sovereignNodeFactory := NewSovereignNodeFactory()

	sn, err := sovereignNodeFactory.CreateNewNode()
	assert.Nil(t, err)
	assert.NotNil(t, sn)
}
