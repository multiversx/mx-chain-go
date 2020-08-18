package delegation

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/stretchr/testify/require"
)

var NewBalance = arwen.NewBalance
var NewBalanceBig = arwen.NewBalanceBig
var RequireAlmostEquals = arwen.RequireAlmostEquals

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

	// Genesis
	context.GasLimit = 30000000
	deployDelegation(context)
	addNodes(context, 2500, 2)

	err := context.ExecuteSC(&context.Alice, "stakeGenesis@"+NewBalance(3000).ToHex())
	require.Nil(t, err)
	err = context.ExecuteSC(&context.Bob, "stakeGenesis@00"+NewBalance(2000).ToHex())
	require.Nil(t, err)

	context.GasLimit = 60000000
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

	// Alice, Bob and Carol claim their rewards
	context.TakeAccountBalanceSnapshot(&context.Alice)
	context.TakeAccountBalanceSnapshot(&context.Bob)
	context.TakeAccountBalanceSnapshot(&context.Carol)

	context.GasLimit = 20000000
	err = context.ExecuteSC(&context.Alice, "claimRewards")
	require.Nil(t, err)
	require.Equal(t, 15518713, int(context.LastConsumedFee))
	RequireAlmostEquals(t, NewBalance(600), NewBalanceBig(context.GetAccountBalanceDelta(&context.Alice)))

	err = context.ExecuteSC(&context.Bob, "claimRewards")
	require.Nil(t, err)
	require.Equal(t, 15077713, int(context.LastConsumedFee))
	RequireAlmostEquals(t, NewBalance(400), NewBalanceBig(context.GetAccountBalanceDelta(&context.Bob)))

	err = context.ExecuteSC(&context.Carol, "claimRewards")
	require.Equal(t, errors.New("user error"), err)
}

func TestDelegation_WithManyUsers_Claims(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	var err error

	stakePerNode := 2500
	numNodes := 1000
	totalStaked := stakePerNode * numNodes
	numUsers := 100
	stakePerUser := totalStaked / numUsers

	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.InitAdditionalParticipants(numUsers)

	// Genesis
	context.GasLimit = 10000000000
	deployDelegation(context)
	addNodes(context, stakePerNode, numNodes)

	for _, user := range context.Participants {
		err = context.ExecuteSC(user, "stakeGenesis@"+NewBalance(stakePerUser).ToHex())
		require.Nil(t, err)
	}

	context.GasLimit = 10000000000
	err = context.ExecuteSC(&context.Owner, "activateGenesis")
	require.Nil(t, err)
	require.Equal(t, numUsers+1, int(context.QuerySCInt("getNumUsers", [][]byte{})))

	// Blockchain continues
	context.GoToEpoch(1)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(5000).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(5000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))

	context.GoToEpoch(2)
	err = context.RewardsProcessor.ProcessRewardTransaction(&rewardTx.RewardTx{Value: NewBalance(5000).Value, RcvAddr: context.ScAddress})
	require.Nil(t, err)
	require.Equal(t, NewBalance(10000).Value, context.QuerySCBigInt("getTotalCumulatedRewards", [][]byte{}))

	// All users claim their rewards
	for _, user := range context.Participants {
		context.TakeAccountBalanceSnapshot(user)

		context.GasLimit = 20000000
		err = context.ExecuteSC(user, "claimRewards")
		require.Nil(t, err)
		require.LessOrEqual(t, int(context.LastConsumedFee), 17000000)
		RequireAlmostEquals(t, NewBalance(10000/numUsers), NewBalanceBig(context.GetAccountBalanceDelta(user)))
	}
}

func deployDelegation(context *arwen.TestContext) {
	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"

	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(context.T, err)
}

func addNodes(context *arwen.TestContext, stakePerNode int, numNodes int) {
	err := context.ExecuteSC(&context.Owner, "setStakePerNode@"+NewBalance(stakePerNode).ToHex())
	require.Nil(context.T, err)

	addNodesArguments := make([]string, 0, numNodes*2)
	for tag := 0; tag < numNodes; tag++ {
		tagBytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(tagBytes, uint32(tag))

		blsKey := bytes.Repeat(tagBytes, 96/len(tagBytes))
		blsSignature := bytes.Repeat(tagBytes, 32/len(tagBytes))

		addNodesArguments = append(addNodesArguments, hex.EncodeToString(blsKey))
		addNodesArguments = append(addNodesArguments, hex.EncodeToString(blsSignature))
	}

	err = context.ExecuteSC(&context.Owner, "addNodes@"+strings.Join(addNodesArguments, "@"))
	require.Nil(context.T, err)
	require.Equal(context.T, numNodes, int(context.QuerySCInt("getNumNodes", [][]byte{})))
	fmt.Println("addNodes consumed (gas):", context.LastConsumedFee)
}
