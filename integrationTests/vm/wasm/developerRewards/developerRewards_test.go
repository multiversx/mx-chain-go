package transfers

import (
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func TestClaimDeveloperRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	wasmPath := "../testdata/developer-rewards/output/developer_rewards.wasm"

	t.Run("rewards for user", func(t *testing.T) {
		context := wasm.SetupTestContext(t)
		defer context.Close()

		err := context.DeploySC(wasmPath, "")
		require.Nil(t, err)
		contractAddress := context.ScAddress

		err = context.ExecuteSC(&context.Owner, "doSomething")
		require.Nil(t, err)

		ownerBalanceBefore := context.GetAccountBalance(&context.Owner).Uint64()
		rewardBig := context.GetAccount(contractAddress).GetDeveloperReward()
		reward := rewardBig.Uint64()

		err = context.ExecuteSC(&context.Owner, "ClaimDeveloperRewards")
		require.Nil(t, err)

		ownerBalanceAfter := context.GetAccountBalance(&context.Owner).Uint64()
		require.Equal(t, ownerBalanceBefore-context.LastConsumedFee+reward, ownerBalanceAfter)

		events := context.LastLogs[0].GetLogEvents()
		require.Equal(t, "ClaimDeveloperRewards", string(events[0].GetIdentifier()))
		require.Equal(t, rewardBig.Bytes(), events[0].GetTopics()[0])
		require.Equal(t, context.Owner.Address, events[0].GetTopics()[1])
	})

	t.Run("rewards for contract", func(t *testing.T) {
		context := wasm.SetupTestContext(t)
		defer context.Close()

		err := context.DeploySC(wasmPath, "")
		require.Nil(t, err)
		parentContractAddress := context.ScAddress

		err = context.ExecuteSC(&context.Owner, "deployChild")
		require.Nil(t, err)

		chilContractdAddress := context.QuerySCBytes("getChildAddress", [][]byte{})
		require.NotNil(t, chilContractdAddress)

		context.ScAddress = chilContractdAddress
		err = context.ExecuteSC(&context.Owner, "doSomething")
		require.Nil(t, err)

		contractBalanceBefore := context.GetAccount(parentContractAddress).GetBalance().Uint64()
		rewardBig := context.GetAccount(chilContractdAddress).GetDeveloperReward()
		reward := rewardBig.Uint64()

		context.ScAddress = parentContractAddress
		err = context.ExecuteSC(&context.Owner, "claimDeveloperRewardsOnChild")
		require.Nil(t, err)

		contractBalanceAfter := context.GetAccount(parentContractAddress).GetBalance().Uint64()
		require.Equal(t, contractBalanceBefore+reward, contractBalanceAfter)

		events := context.LastLogs[0].GetLogEvents()
		require.Equal(t, "ClaimDeveloperRewards", string(events[0].GetIdentifier()))
		require.Equal(t, rewardBig.Bytes(), events[0].GetTopics()[0])
		require.Equal(t, parentContractAddress, events[0].GetTopics()[1])
	})
}
