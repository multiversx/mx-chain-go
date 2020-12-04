package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStakingDataProvider_NilSystemVMShouldErr(t *testing.T) {
	t.Parallel()

	sdp, err := NewStakingDataProvider(nil, "100000")

	assert.True(t, check.IfNil(sdp))
	assert.Equal(t, epochStart.ErrNilSystemVmInstance, err)
}

func TestNewStakingDataProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	sdp, err := NewStakingDataProvider(&mock.VMExecutionHandlerStub{}, "100000")

	assert.False(t, check.IfNil(sdp))
	assert.Nil(t, err)
}

func TestStakingDataProvider_GPrepareDataForBlsKeyGetBlsKeyOwnerErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	expectedErr := errors.New("expected error")
	sdp, _ := NewStakingDataProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			numCall++
			if numCall == 1 {
				return nil, expectedErr
			}
			if numCall == 2 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				}, nil
			}
			if numCall == 3 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
				}, nil
			}

			return nil, nil
		},
	}, "100000")

	err := sdp.loadDataForBlsKey([]byte("bls key"))
	assert.Equal(t, expectedErr, err)

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "returned exactly one value: the owner address"))
}

func TestStakingDataProvider_PrepareDataForBlsKeyLoadOwnerDataErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	owner := []byte("owner")
	expectedErr := errors.New("expected error")
	sdp, _ := NewStakingDataProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			if input.Function == "getOwner" {
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(hex.EncodeToString(owner))},
				}, nil
			}

			numCall++
			if numCall == 1 {
				return nil, expectedErr
			}
			if numCall == 2 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.UserError,
				}, nil
			}
			if numCall == 3 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
				}, nil
			}
			if numCall == 4 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
					ReturnData: [][]byte{[]byte("not a number"), []byte("1")},
				}, nil
			}
			if numCall == 5 {
				return &vmcommon.VMOutput{
					ReturnCode: vmcommon.Ok,
					ReturnData: [][]byte{[]byte("1"), []byte("not a number")},
				}, nil
			}

			return nil, nil
		},
	}, "100000")

	err := sdp.loadDataForBlsKey([]byte("bls key"))
	assert.Equal(t, expectedErr, err)

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "getTotalStakedTopUpBlsKeys function should have at least two values"))

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "topUp string returned is not a number"))

	err = sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "totalStaked string returned is not a number"))
}

func TestStakingDataProvider_PrepareDataForBlsKeyReturnedOwnerIsNotHexShouldErr(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	sdp, _ := NewStakingDataProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			if input.Function == "getOwner" {
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{owner},
				}, nil
			}

			return nil, nil
		},
	}, "100000")

	err := sdp.loadDataForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
}

func TestStakingDataProvider_PrepareDataForBlsKeyFromSCShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	err := sdp.loadDataForBlsKey([]byte("bls key"))
	assert.Nil(t, err)
	assert.Equal(t, 2, numRunContractCalls)
	ownerData := sdp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 1, ownerData.numEligible)
}

func TestStakingDataProvider_PrepareDataForBlsKeyCachedResponseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	basePrice := big.NewInt(100000)
	stakeVal := big.NewInt(0).Add(topUpVal, basePrice)
	numRunContractCalls := 0

	sdp := createStakingDataProviderWithMockArgs(t, owner, topUpVal, stakeVal, &numRunContractCalls)

	err := sdp.loadDataForBlsKey([]byte("bls key"))
	assert.Nil(t, err)

	err = sdp.loadDataForBlsKey([]byte("bls key2"))
	assert.Nil(t, err)

	assert.Equal(t, 3, numRunContractCalls)
	ownerData := sdp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 2, ownerData.numEligible)
}

func TestStakingDataProvider_PrepareDataForBlsKeyWithRealSystemVmShouldWork(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	sdp := createStakingDataProviderWithRealArgs(t, owner, blsKey, topUpVal)
	err := sdp.loadDataForBlsKey(blsKey)
	assert.Nil(t, err)
	ownerData := sdp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 1, ownerData.numEligible)
}

func createStakingDataProviderWithMockArgs(
	t *testing.T,
	owner []byte,
	topUpVal *big.Int,
	stakingVal *big.Int,
	numRunContractCalls *int,
) *stakingDataProvider {
	sdp, err := NewStakingDataProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			*numRunContractCalls++
			switch input.Function {
			case "getOwner":
				assert.Equal(t, vm.ValidatorSCAddress, input.VMInput.CallerAddr)
				assert.Equal(t, vm.StakingSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(hex.EncodeToString(owner))},
				}, nil
			case "getTotalStakedTopUpBlsKeys":
				assert.Equal(t, owner, input.VMInput.CallerAddr)
				assert.Equal(t, vm.ValidatorSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(topUpVal.String()), []byte(stakingVal.String())},
				}, nil

			}

			return nil, errors.New("unexpected call")
		},
	}, "100000")
	require.Nil(t, err)

	return sdp
}

func createStakingDataProviderWithRealArgs(t *testing.T, owner []byte, blsKey []byte, topUpVal *big.Int) *stakingDataProvider {
	args, _ := createFullArgumentsForSystemSCProcessing(1000)
	args.EpochNotifier.CheckEpoch(1000000)
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	doStake(t, s.systemVM, s.userAccountsDB, owner, big.NewInt(0).Add(big.NewInt(1000), topUpVal), blsKey)

	sdp, _ := NewStakingDataProvider(s.systemVM, "100000")

	return sdp
}

func saveOutputAccounts(t *testing.T, accountsDB state.AccountsAdapter, vmOutput *vmcommon.VMOutput) {
	for _, outputAccount := range vmOutput.OutputAccounts {
		account, errLoad := accountsDB.LoadAccount(outputAccount.Address)
		require.Nil(t, errLoad)

		userAccount, _ := account.(state.UserAccountHandler)
		for _, storeUpdate := range outputAccount.StorageUpdates {
			_ = userAccount.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		err := accountsDB.SaveAccount(account)
		require.Nil(t, err)
	}

	_, err := accountsDB.Commit()
	require.Nil(t, err)
}
