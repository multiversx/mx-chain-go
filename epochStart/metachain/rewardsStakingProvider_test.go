package metachain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRewardsStakingProvider_NilSystemVMShouldErr(t *testing.T) {
	t.Parallel()

	rsp, err := NewRewardsStakingProvider(nil)

	assert.True(t, check.IfNil(rsp))
	assert.Equal(t, epochStart.ErrNilSystemVmInstance, err)
}

func TestNewRewardsStakingProvider_ShouldWork(t *testing.T) {
	t.Parallel()

	rsp, err := NewRewardsStakingProvider(&mock.VMExecutionHandlerStub{})

	assert.False(t, check.IfNil(rsp))
	assert.Nil(t, err)
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyGetBlsKeyOwnerErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	expectedErr := errors.New("expected error")
	rsp, _ := NewRewardsStakingProvider(&mock.VMExecutionHandlerStub{
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
	})

	err := rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.Equal(t, expectedErr, err)

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "missing owner address"))
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyLoadOwnerDataErrorsShouldErr(t *testing.T) {
	t.Parallel()

	numCall := 0
	owner := []byte("owner")
	expectedErr := errors.New("expected error")
	rsp, _ := NewRewardsStakingProvider(&mock.VMExecutionHandlerStub{
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
					ReturnData: [][]byte{[]byte("not a number")},
				}, nil
			}

			return nil, nil
		},
	})

	err := rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.Equal(t, expectedErr, err)

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), vmcommon.UserError.String()))

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "missing top up value data bytes"))

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), epochStart.ErrExecutingSystemScCode.Error()))
	assert.True(t, strings.Contains(err.Error(), "topUp string returned is not a number"))
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyReturnedOwnerIsNotHexShouldErr(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	rsp, _ := NewRewardsStakingProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			if input.Function == "getOwner" {
				return &vmcommon.VMOutput{
					ReturnData: [][]byte{owner},
				}, nil
			}

			return nil, nil
		},
	})

	err := rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.NotNil(t, err)
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyFromSCShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	numRunContractCalls := 0

	rsp := createRewardsStakingProviderWithMockArgs(t, owner, topUpVal, &numRunContractCalls)

	err := rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.Nil(t, err)
	assert.Equal(t, 2, numRunContractCalls)
	ownerData := rsp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 1, ownerData.numEligible)
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyCachedResponseShouldWork(t *testing.T) {
	t.Parallel()

	owner := []byte("owner")
	topUpVal := big.NewInt(828743)
	numRunContractCalls := 0

	rsp := createRewardsStakingProviderWithMockArgs(t, owner, topUpVal, &numRunContractCalls)

	err := rsp.ComputeRewardsForBlsKey([]byte("bls key"))
	assert.Nil(t, err)

	err = rsp.ComputeRewardsForBlsKey([]byte("bls key2"))
	assert.Nil(t, err)

	assert.Equal(t, 3, numRunContractCalls)
	ownerData := rsp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 2, ownerData.numEligible)
}

func TestRewardsStakingProvider_ComputeRewardsForBlsKeyWithRealSystemVmShouldWork(t *testing.T) {
	t.Parallel()

	owner := append([]byte("owner"), bytes.Repeat([]byte{1}, 27)...)
	topUpVal := big.NewInt(828743)
	blsKey := []byte("bls key")

	rsp := createRewardsStakingProviderWithRealArgs(t, owner, blsKey, topUpVal)
	err := rsp.ComputeRewardsForBlsKey(blsKey)
	assert.Nil(t, err)
	ownerData := rsp.GetFromCache([]byte(hex.EncodeToString(owner)))
	require.NotNil(t, ownerData)
	assert.Equal(t, topUpVal, ownerData.topUpValue)
	assert.Equal(t, 1, ownerData.numEligible)
}

func createRewardsStakingProviderWithMockArgs(t *testing.T, owner []byte, topUpVal *big.Int, numRunContractCalls *int) *rewardsStakingProvider {
	rsp, _ := NewRewardsStakingProvider(&mock.VMExecutionHandlerStub{
		RunSmartContractCallCalled: func(input *vmcommon.ContractCallInput) (*vmcommon.VMOutput, error) {
			*numRunContractCalls++
			switch input.Function {
			case "getOwner":
				assert.Equal(t, vm.AuctionSCAddress, input.VMInput.CallerAddr)
				assert.Equal(t, vm.StakingSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(hex.EncodeToString(owner))},
				}, nil
			case "getTopUp":
				assert.Equal(t, owner, input.VMInput.CallerAddr)
				assert.Equal(t, vm.AuctionSCAddress, input.RecipientAddr)

				return &vmcommon.VMOutput{
					ReturnData: [][]byte{[]byte(topUpVal.String())},
				}, nil
			}

			return nil, errors.New("unexpected call")
		},
	})

	return rsp
}

func createRewardsStakingProviderWithRealArgs(t *testing.T, owner []byte, blsKey []byte, topUpVal *big.Int) *rewardsStakingProvider {
	args := createFullArgumentsForSystemSCProcessing()
	args.EpochNotifier.CheckEpoch(1000000)
	s, _ := NewSystemSCProcessor(args)
	require.NotNil(t, s)

	doStake(t, s.systemVM, s.userAccountsDB, owner, big.NewInt(0).Add(big.NewInt(1000), topUpVal), blsKey)

	rsp, _ := NewRewardsStakingProvider(s.systemVM)

	return rsp
}

func saveOutputAccounts(t *testing.T, accountsDB state.AccountsAdapter, vmOutput *vmcommon.VMOutput) {
	for _, outputAccount := range vmOutput.OutputAccounts {
		account, errLoad := accountsDB.LoadAccount(outputAccount.Address)
		require.Nil(t, errLoad)

		userAccount, _ := account.(state.UserAccountHandler)
		for _, storeUpdate := range outputAccount.StorageUpdates {
			userAccount.DataTrieTracker().SaveKeyValue(storeUpdate.Offset, storeUpdate.Data)
		}

		err := accountsDB.SaveAccount(account)
		require.Nil(t, err)
	}

	_, err := accountsDB.Commit()
	require.Nil(t, err)
}
