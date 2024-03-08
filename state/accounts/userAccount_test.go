package accounts_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/common/errChan"
	"github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/state/accounts"
	testTrie "github.com/multiversx/mx-chain-go/testscommon/trie"
	"github.com/stretchr/testify/assert"
)

func createUserAcc(address []byte) state.UserAccountHandler {
	acc, _ := accounts.NewUserAccount(address, &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
	return acc
}

func TestNewUserAccount(t *testing.T) {
	t.Parallel()

	t.Run("nil address", func(t *testing.T) {
		t.Parallel()

		acc, err := accounts.NewUserAccount(nil, &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, errors.ErrNilAddress, err)
	})

	t.Run("nil data trie tracker", func(t *testing.T) {
		t.Parallel()

		acc, err := accounts.NewUserAccount(make([]byte, 32), nil, &testTrie.TrieLeafParserStub{})
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, errors.ErrNilTrackableDataTrie, err)
	})

	t.Run("nil trie leaf parser", func(t *testing.T) {
		t.Parallel()

		acc, err := accounts.NewUserAccount(make([]byte, 32), &testTrie.DataTrieTrackerStub{}, nil)
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, errors.ErrNilTrieLeafParser, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		acc, err := accounts.NewUserAccount(make([]byte, 32), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
		assert.Nil(t, err)
		assert.False(t, check.IfNil(acc))
	})
}

func TestUserAccount_AddressBytes(t *testing.T) {
	t.Parallel()

	address := []byte("address bytes")
	acc := createUserAcc(address)

	assert.Equal(t, address, acc.AddressBytes())
}

func TestUserAccount_AddToBalanceInsufficientFundsShouldErr(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	value := big.NewInt(-1)

	err := acc.AddToBalance(value)
	assert.Equal(t, state.ErrInsufficientFunds, err)
}

func TestUserAccount_SubFromBalanceInsufficientFundsShouldErr(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	value := big.NewInt(1)

	err := acc.SubFromBalance(value)
	assert.Equal(t, state.ErrInsufficientFunds, err)
}

func TestUserAccount_GetBalance(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	balance := big.NewInt(100)
	subFromBalance := big.NewInt(20)

	_ = acc.AddToBalance(balance)
	assert.Equal(t, balance, acc.GetBalance())
	_ = acc.SubFromBalance(subFromBalance)
	assert.Equal(t, big.NewInt(0).Sub(balance, subFromBalance), acc.GetBalance())
}

func TestUserAccount_AddToDeveloperReward(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	reward := big.NewInt(10)

	acc.AddToDeveloperReward(reward)
	assert.Equal(t, reward, acc.GetDeveloperReward())
}

func TestUserAccount_ClaimDeveloperRewardsWrongAddressShouldErr(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	val, err := acc.ClaimDeveloperRewards([]byte("wrong address"))
	assert.Nil(t, val)
	assert.Equal(t, state.ErrOperationNotPermitted, err)
}

func TestUserAccount_ClaimDeveloperRewards(t *testing.T) {
	t.Parallel()

	acc, _ := accounts.NewUserAccount(make([]byte, 32), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
	reward := big.NewInt(10)
	acc.AddToDeveloperReward(reward)

	val, err := acc.ClaimDeveloperRewards(acc.OwnerAddress)
	assert.Nil(t, err)
	assert.Equal(t, reward, val)
	assert.Equal(t, big.NewInt(0), acc.GetDeveloperReward())
}

func TestUserAccount_ChangeOwnerAddressWrongAddressShouldErr(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	err := acc.ChangeOwnerAddress([]byte("wrong address"), []byte{})
	assert.Equal(t, state.ErrOperationNotPermitted, err)
}

func TestUserAccount_ChangeOwnerAddressInvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	acc, _ := accounts.NewUserAccount(make([]byte, 32), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
	err := acc.ChangeOwnerAddress(acc.OwnerAddress, []byte("new address"))
	assert.Equal(t, errors.ErrInvalidAddressLength, err)
}

func TestUserAccount_ChangeOwnerAddress(t *testing.T) {
	t.Parallel()

	newAddress := make([]byte, 32)
	acc, _ := accounts.NewUserAccount(make([]byte, 32), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})

	err := acc.ChangeOwnerAddress(acc.OwnerAddress, newAddress)
	assert.Nil(t, err)
	assert.Equal(t, newAddress, acc.GetOwnerAddress())
}

func TestUserAccount_SetOwnerAddress(t *testing.T) {
	t.Parallel()

	newAddress := []byte("new address")
	acc := createUserAcc(make([]byte, 32))

	acc.SetOwnerAddress(newAddress)
	assert.Equal(t, newAddress, acc.GetOwnerAddress())
}

func TestUserAccount_SetAndGetNonce(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	nonce := uint64(5)

	acc.IncreaseNonce(nonce)
	assert.Equal(t, nonce, acc.GetNonce())
}

func TestUserAccount_SetAndGetCodeHash(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	codeHash := []byte("code hash")

	acc.SetCodeHash(codeHash)
	assert.Equal(t, codeHash, acc.GetCodeHash())
}

func TestUserAccount_SetAndGetRootHash(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	rootHash := []byte("root hash")

	acc.SetRootHash(rootHash)
	assert.Equal(t, rootHash, acc.GetRootHash())
}

func TestUserAccount_SetAndGetVersion(t *testing.T) {
	t.Parallel()

	acc := createUserAcc(make([]byte, 32))
	version := uint8(2)

	acc.SetVersion(version)
	assert.Equal(t, version, acc.GetVersion())
}

func TestUserAccount_GetAllLeaves(t *testing.T) {
	t.Parallel()

	t.Run("nil data trie should err", func(t *testing.T) {
		t.Parallel()

		acc, _ := accounts.NewUserAccount([]byte("address"), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		err := acc.GetAllLeaves(chLeaves, nil)
		assert.Equal(t, errors.ErrNilTrie, err)
	})

	t.Run("can not retrieve root hash should err", func(t *testing.T) {
		t.Parallel()

		expectedErr := fmt.Errorf("root error")
		dtt := &testTrie.DataTrieTrackerStub{
			DataTrieCalled: func() common.Trie {
				return &testTrie.TrieStub{
					RootCalled: func() ([]byte, error) {
						return nil, expectedErr
					},
				}
			},
		}

		acc, _ := accounts.NewUserAccount([]byte("address"), dtt, &testTrie.TrieLeafParserStub{})
		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		err := acc.GetAllLeaves(chLeaves, nil)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("should call GetAllLeavesOnChannel from trie", func(t *testing.T) {
		t.Parallel()

		dtlp := &testTrie.TrieLeafParserStub{}
		rootHash := []byte("root hash")
		getAllLeavesCalled := false
		dtt := &testTrie.DataTrieTrackerStub{
			DataTrieCalled: func() common.Trie {
				return &testTrie.TrieStub{
					RootCalled: func() ([]byte, error) {
						return rootHash, nil
					},
					GetAllLeavesOnChannelCalled: func(_ *common.TrieIteratorChannels, _ context.Context, hash []byte, _ common.KeyBuilder, trieLeafParser common.TrieLeafParser) error {
						getAllLeavesCalled = true
						assert.Equal(t, rootHash, hash)
						assert.Equal(t, dtlp, trieLeafParser)

						return nil
					},
				}
			},
		}

		acc, _ := accounts.NewUserAccount([]byte("address"), dtt, dtlp)
		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    errChan.NewErrChanWrapper(),
		}

		err := acc.GetAllLeaves(chLeaves, nil)
		assert.Nil(t, err)
		assert.True(t, getAllLeavesCalled)
	})
}

func TestUserAccount_IsDataTrieMigrated(t *testing.T) {
	t.Parallel()

	t.Run("nil trie", func(t *testing.T) {
		t.Parallel()

		acc, _ := accounts.NewUserAccount([]byte("address"), &testTrie.DataTrieTrackerStub{}, &testTrie.TrieLeafParserStub{})
		isMigrated, err := acc.IsDataTrieMigrated()
		assert.False(t, isMigrated)
		assert.Equal(t, state.ErrNilTrie, err)
	})

	t.Run("calls IsMigratedToLatestVersion from trie", func(t *testing.T) {
		t.Parallel()

		isMigratedCalled := false
		dtt := &testTrie.DataTrieTrackerStub{
			DataTrieCalled: func() common.Trie {
				return &testTrie.TrieStub{
					IsMigratedToLatestVersionCalled: func() (bool, error) {
						isMigratedCalled = true
						return false, nil
					},
				}
			},
		}

		acc, _ := accounts.NewUserAccount([]byte("address"), dtt, &testTrie.TrieLeafParserStub{})

		isMigrated, err := acc.IsDataTrieMigrated()
		assert.False(t, isMigrated)
		assert.Nil(t, err)
		assert.True(t, isMigratedCalled)
	})
}
