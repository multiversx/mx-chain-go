package delegation

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
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
	err = context.ExecuteSC(&context.Owner, "setStakePerNode@"+NewBalance(2500).ToHex())
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "addNodes@"+createAddNodesParam('a')+"@"+createAddNodesParam('b'))
	require.Nil(t, err)
	require.Equal(t, 2, int(context.QuerySCInt("getNumNodes", [][]byte{})))

	err = context.ExecuteSC(&context.Alice, "stakeGenesis@"+NewBalance(3000).ToHex())
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Bob, "stakeGenesis@00"+NewBalance(2000).ToHex())
	require.Nil(t, err)

	err = context.ExecuteSC(&context.Owner, "activateGenesis")
	require.Nil(t, err)

	require.Equal(t, 3, int(context.QuerySCInt("getNumUsers", [][]byte{})))
	require.Equal(t, NewBalance(3000).Value, context.QuerySCBigInt("getUserStake", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(2000).Value, context.QuerySCBigInt("getUserStake", [][]byte{context.Bob.Address}))

	// No rewards yet
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))
	require.Equal(t, big.NewInt(0), context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(5000).Value, context.QuerySCBigInt("getTotalActiveStake", [][]byte{}))

	// Blockchain continues
	context.GoToEpoch(1)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(500).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(500).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(300).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(200).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	context.GoToEpoch(2)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(500).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(1000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))
	require.Equal(t, NewBalance(600).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Alice.Address}))
	require.Equal(t, NewBalance(400).Value, context.QuerySCBigInt("getClaimableRewards", [][]byte{context.Bob.Address}))

	err = context.ExecuteSC(&context.Carol, "claimRewards")
	require.Equal(t, errors.New("user error"), err)

	context.TakeAccountBalanceSnapshot(&context.Alice)
	context.TakeAccountBalanceSnapshot(&context.Bob)

	err = context.ExecuteSC(&context.Alice, "claimRewards")
	require.Nil(t, err)
	RequireAlmostEquals(t, NewBalance(600), NewBalanceBig(context.GetAccountBalanceDelta(&context.Alice)))

	err = context.ExecuteSC(&context.Bob, "claimRewards")
	require.Nil(t, err)
	RequireAlmostEquals(t, NewBalance(400), NewBalanceBig(context.GetAccountBalanceDelta(&context.Bob)))
}

func createAddNodesParam(tag byte) string {
	blsKey := bytes.Repeat([]byte{tag}, 96)
	blsSignature := bytes.Repeat([]byte{tag}, 32)
	return hex.EncodeToString(blsKey) + "@" + hex.EncodeToString(blsSignature)
}

// Balance -
type Balance struct {
	Value *big.Int
}

// NewBalance
func NewBalance(n int) Balance {
	result := big.NewInt(0)
	_, _ = result.SetString("1000000000000000000", 10)
	result.Mul(result, big.NewInt(int64(n)))
	return Balance{Value: result}
}

// NewBalanceBig
func NewBalanceBig(bi *big.Int) Balance {
	return Balance{Value: bi}
}

// Times -
func (b Balance) Times(n int) Balance {
	result := b.Value.Mul(b.Value, big.NewInt(int64(n)))
	return Balance{Value: result}
}

// ToHex -
func (b Balance) ToHex() string {
	return "00" + hex.EncodeToString(b.Value.Bytes())
}

// AlmostEquals -
func RequireAlmostEquals(t *testing.T, expected Balance, actual Balance) {
	precision := big.NewInt(0)
	_, _ = precision.SetString("100000000000", 10)
	delta := big.NewInt(0)
	delta = delta.Sub(expected.Value, actual.Value)
	delta = delta.Abs(delta)
	require.True(t, delta.Cmp(precision) < 0, fmt.Sprintf("%s != %s", expected, actual))
}
