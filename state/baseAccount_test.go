package state_test

import (
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/state"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func TestBaseAccount_AddressContainer(t *testing.T) {
	t.Parallel()

	address := make([]byte, 32)

	ba := state.NewEmptyBaseAccount(address, nil)
	assert.Equal(t, address, ba.AddressBytes())
}

func TestBaseAccount_DataTrie(t *testing.T) {
	t.Parallel()

	tr := &trieMock.TrieStub{}
	setCalled := false
	getCalled := false

	tracker := &trieMock.DataTrieTrackerStub{
		SetDataTrieCalled: func(tr common.Trie) {
			setCalled = true
		},
		DataTrieCalled: func() common.Trie {
			getCalled = true
			return tr
		},
	}

	ba := state.NewEmptyBaseAccount(nil, tracker)
	ba.SetDataTrie(tr)

	assert.Equal(t, tr, ba.DataTrie())
	assert.True(t, setCalled)
	assert.True(t, getCalled)
}

func TestBaseAccount_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	ba := state.NewEmptyBaseAccount(nil, nil)
	assert.False(t, check.IfNil(ba))
	ba = nil
	assert.True(t, check.IfNil(ba))
}
