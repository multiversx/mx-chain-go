package delegation

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

func TestDelegation_Upgrade(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	delegationWasmPathA := "../testdata/delegation/delegation_vA.wasm"
	delegationWasmPathB := "../testdata/delegation/delegation_vB.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@0080@00@0080@0080"
	delegationUpgradeParams := "0000000000000000000000000000000000000000000000000000000000000000@0080@00@0080@0080"

	context.ScCodeMetadata.Upgradeable = true
	context.GasLimit = 40000000

	err := context.DeploySC(delegationWasmPathA, delegationInitParams)
	require.Nil(t, err)
	account, err := context.Accounts.GetExistingAccount(context.ScAddress)
	require.Nil(t, err)
	codeHashA := account.(state.UserAccountHandler).GetCodeHash()

	context.GasLimit = 21700000
	err = context.UpgradeSC(delegationWasmPathB, delegationUpgradeParams)
	require.Nil(t, err)
	account, err = context.Accounts.GetExistingAccount(context.ScAddress)
	require.Nil(t, err)
	codeHashB := account.(state.UserAccountHandler).GetCodeHash()

	require.NotEqual(t, codeHashA, codeHashB)
}

func TestDelegation_Claims(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"

	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(t, err)

	// Genesis
	err = context.ExecuteSC(&context.Owner, "setStakePerNode@"+hex.EncodeToString(big.NewInt(1000000).Bytes()))
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "addNodes@"+createAddNodesParam('a')+"@"+createAddNodesParam('b'))
	require.Nil(t, err)
	require.Equal(t, 2, int(context.QuerySCInt("getNumNodes", [][]byte{})))

	err = context.ExecuteSC(&context.Alice, "stakeGenesis@00"+hex.EncodeToString(big.NewInt(1000000).Bytes()))
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Bob, "stakeGenesis@00"+hex.EncodeToString(big.NewInt(1000000).Bytes()))
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "activateGenesis")
	require.Nil(t, err)

	require.Equal(t, 3, int(context.QuerySCInt("getNumUsers", [][]byte{})))
	require.Equal(t, big.NewInt(1000000), context.QuerySCBigInt("getUserStake", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(1000000), context.QuerySCBigInt("getUserStake", [][]byte{context.Bob.Address}))

	// No rewards yet
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, big.NewInt(2000000), context.QuerySCBigInt("getTotalActiveStake", [][]byte{}))

	// Blockchain continues
	context.GoToEpoch(1)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: big.NewInt(1000), RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, big.NewInt(1000), context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, big.NewInt(500), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(500), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	context.GoToEpoch(2)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: big.NewInt(1000), RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, big.NewInt(2000), context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, big.NewInt(1000), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(1000), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	err = context.ExecuteSC(&context.Carol, "claimRewards")
	require.Equal(t, errors.New("user error"), err)

	err = context.ExecuteSC(&context.Alice, "claimRewards")
	require.Nil(t, err)

	// TODO: assert balances
	// TODO: add tests for many users (delegators), check gas limit
}

func createAddNodesParam(tag byte) string {
	blsKey := bytes.Repeat([]byte{tag}, 96)
	blsSignature := bytes.Repeat([]byte{tag}, 32)
	return hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(blsSignature)
}
