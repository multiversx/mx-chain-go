package state_test

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/enableEpochsHandlerMock"
	"github.com/stretchr/testify/assert"
)

func TestNewUserAccount(t *testing.T) {
	t.Parallel()

	t.Run("nil address", func(t *testing.T) {
		t.Parallel()

		acc, err := state.NewUserAccount(nil, getDefaultArgsAccountCreation())
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, state.ErrNilAddress, err)
	})

	t.Run("nil hasher", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgsAccountCreation()
		args.Hasher = nil
		acc, err := state.NewUserAccount(make([]byte, 32), args)
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, state.ErrNilHasher, err)
	})

	t.Run("nil marshaller", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgsAccountCreation()
		args.Marshaller = nil
		acc, err := state.NewUserAccount(make([]byte, 32), args)
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, state.ErrNilMarshalizer, err)
	})

	t.Run("nil enableEpochsHandler", func(t *testing.T) {
		t.Parallel()

		args := getDefaultArgsAccountCreation()
		args.EnableEpochsHandler = nil
		acc, err := state.NewUserAccount(make([]byte, 32), args)
		assert.True(t, check.IfNil(acc))
		assert.Equal(t, state.ErrNilEnableEpochsHandler, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		acc, err := state.NewUserAccount(make([]byte, 32), getDefaultArgsAccountCreation())
		assert.Nil(t, err)
		assert.False(t, check.IfNil(acc))
	})
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

	acc, _ := state.NewUserAccount(make([]byte, 32), getDefaultArgsAccountCreation())
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

	acc, _ := state.NewUserAccount(make([]byte, 32), getDefaultArgsAccountCreation())
	err := acc.ChangeOwnerAddress(acc.OwnerAddress, []byte("new address"))
	assert.Equal(t, state.ErrInvalidAddressLength, err)
}

func TestUserAccount_ChangeOwnerAddress(t *testing.T) {
	t.Parallel()

	newAddress := make([]byte, 32)
	acc, _ := state.NewUserAccount(make([]byte, 32), getDefaultArgsAccountCreation())

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

func TestUserAccount_GetAllLeaves(t *testing.T) {
	t.Parallel()

	t.Run("autoBalance data tries disabled", func(t *testing.T) {
		t.Parallel()

		tr, _ := getDefaultTrieAndAccountsDb()
		acc, _ := state.NewUserAccount([]byte("address"), getDefaultArgsAccountCreation())
		numKeys := 1000
		vals := make(map[string][]byte)
		for i := 0; i < numKeys; i++ {
			key := []byte(strconv.Itoa(i))
			val := []byte(strconv.Itoa(i))
			vals[string(key)] = val
			err := acc.SaveKeyValue(key, val)
			assert.Nil(t, err)
		}
		acc.SetDataTrie(tr)
		_, _ = acc.SaveDirtyData(tr)
		rh, _ := acc.DataTrie().RootHash()
		acc.SetRootHash(rh)
		_ = tr.Commit()

		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    make(chan error, 1),
		}
		err := acc.GetAllLeaves(chLeaves, context.Background())
		assert.Nil(t, err)

		for leaf := range chLeaves.LeavesChan {
			val, ok := vals[string(leaf.Key())]
			assert.True(t, ok)
			assert.Equal(t, val, leaf.Value())
		}
	})

	t.Run("autoBalance data tries enabled", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: true,
		}
		tr, _ := getDefaultTrieAndAccountsDb()
		args := getDefaultArgsAccountCreation()
		args.EnableEpochsHandler = enableEpochsHandler
		acc, _ := state.NewUserAccount([]byte("address"), args)
		numKeys := 1000
		vals := make(map[string][]byte)
		for i := 0; i < numKeys; i++ {
			key := []byte(strconv.Itoa(i))
			val := []byte(strconv.Itoa(i))
			vals[string(key)] = val
			err := acc.SaveKeyValue(key, val)
			assert.Nil(t, err)
		}
		acc.SetDataTrie(tr)
		_, _ = acc.SaveDirtyData(tr)
		rh, _ := acc.DataTrie().RootHash()
		acc.SetRootHash(rh)
		_ = tr.Commit()

		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    make(chan error, 1),
		}
		err := acc.GetAllLeaves(chLeaves, context.Background())
		assert.Nil(t, err)

		for leaf := range chLeaves.LeavesChan {
			val, ok := vals[string(leaf.Key())]
			assert.True(t, ok)
			assert.Equal(t, val, leaf.Value())
		}
	})

	t.Run("autoBalance data tries enabled after insert", func(t *testing.T) {
		t.Parallel()

		enableEpochsHandler := &enableEpochsHandlerMock.EnableEpochsHandlerStub{
			IsAutoBalanceDataTriesEnabledField: false,
		}
		tr, _ := getDefaultTrieAndAccountsDb()
		args := getDefaultArgsAccountCreation()
		args.EnableEpochsHandler = enableEpochsHandler
		acc, _ := state.NewUserAccount([]byte("address"), args)
		numKeys := 1000
		vals := make(map[string][]byte)
		for i := 0; i < numKeys; i++ {
			key := []byte(strconv.Itoa(i))
			val := []byte(strconv.Itoa(i))
			vals[string(key)] = val
			err := acc.SaveKeyValue(key, val)
			assert.Nil(t, err)
		}
		acc.SetDataTrie(tr)
		_, _ = acc.SaveDirtyData(tr)
		rh, _ := acc.DataTrie().RootHash()
		acc.SetRootHash(rh)
		_ = tr.Commit()
		enableEpochsHandler.IsAutoBalanceDataTriesEnabledField = true

		chLeaves := &common.TrieIteratorChannels{
			LeavesChan: make(chan core.KeyValueHolder, 100),
			ErrChan:    make(chan error, 1),
		}
		err := acc.GetAllLeaves(chLeaves, context.Background())
		assert.Nil(t, err)

		for leaf := range chLeaves.LeavesChan {
			val, ok := vals[string(leaf.Key())]
			assert.True(t, ok)
			assert.Equal(t, val, leaf.Value())
		}
	})
}
