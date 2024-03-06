package accounts

import (
	"math/big"
	"testing"

	errorsMx "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/testscommon/sovereign"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignAccount(t *testing.T) {
	t.Parallel()

	t.Run("nil address, should return error", func(t *testing.T) {
		sovAcc, err := NewSovereignAccount(
			nil,
			&trie.DataTrieTrackerStub{},
			&trie.TrieLeafParserStub{},
			&sovereign.ESDTAsBalanceHandlerMock{},
		)
		require.Nil(t, sovAcc)
		require.Equal(t, errorsMx.ErrNilAddress, err)
	})
	t.Run("nil data trie, should return error", func(t *testing.T) {
		sovAcc, err := NewSovereignAccount(
			[]byte("address"),
			nil,
			&trie.TrieLeafParserStub{},
			&sovereign.ESDTAsBalanceHandlerMock{},
		)
		require.Nil(t, sovAcc)
		require.Equal(t, errorsMx.ErrNilTrackableDataTrie, err)
	})
	t.Run("nil leaf parser, should return error", func(t *testing.T) {
		sovAcc, err := NewSovereignAccount(
			[]byte("address"),
			&trie.DataTrieTrackerStub{},
			nil,
			&sovereign.ESDTAsBalanceHandlerMock{},
		)
		require.Nil(t, sovAcc)
		require.Equal(t, errorsMx.ErrNilTrieLeafParser, err)
	})
	t.Run("nil esdt balance, should return error", func(t *testing.T) {
		sovAcc, err := NewSovereignAccount(
			[]byte("address"),
			&trie.DataTrieTrackerStub{},
			&trie.TrieLeafParserStub{},
			nil,
		)
		require.Nil(t, sovAcc)
		require.Equal(t, errorsMx.ErrNilESDTAsBalanceHandler, err)
	})
	t.Run("should work", func(t *testing.T) {
		sovAcc, err := NewSovereignAccount(
			[]byte("address"),
			&trie.DataTrieTrackerStub{},
			&trie.TrieLeafParserStub{},
			&sovereign.ESDTAsBalanceHandlerMock{},
		)
		require.Nil(t, err)
		require.False(t, sovAcc.IsInterfaceNil())
	})
}

func TestSovereignAccount_AddToBalance_SubFromBalance_GetBalance(t *testing.T) {
	t.Parallel()

	wasGetBalanceCalled := false
	wasAddBalanceCalled := false
	wasSubBalanceCalled := false
	esdtBalance := &sovereign.ESDTAsBalanceHandlerMock{
		GetBalanceCalled: func(accountDataHandler vmcommon.AccountDataHandler) *big.Int {
			wasGetBalanceCalled = true
			return nil
		},
		AddToBalanceCalled: func(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
			wasAddBalanceCalled = true
			return nil
		},
		SubFromBalanceCalled: func(accountDataHandler vmcommon.AccountDataHandler, value *big.Int) error {
			wasSubBalanceCalled = true
			return nil
		},
	}
	sovAcc, _ := NewSovereignAccount(
		[]byte("address"),
		&trie.DataTrieTrackerStub{},
		&trie.TrieLeafParserStub{},
		esdtBalance,
	)

	require.Nil(t, sovAcc.GetBalance())
	require.Nil(t, sovAcc.AddToBalance(nil))
	require.Nil(t, sovAcc.SubFromBalance(nil))

	require.True(t, wasGetBalanceCalled)
	require.True(t, wasAddBalanceCalled)
	require.True(t, wasSubBalanceCalled)
}
