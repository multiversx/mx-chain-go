package transfers

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/integrationTests/vm/wasm"
	"github.com/stretchr/testify/require"
)

func TestClaimDeveloperRewards(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	context := wasm.SetupTestContext(t)
	defer context.Close()

	err := context.DeploySC("../testdata/developer-rewards/output/developer_rewards.wasm", "")
	require.Nil(t, err)

	t.Run("rewards for user", func(t *testing.T) {
		contractAddress := context.ScAddress

		err = context.ExecuteSC(&context.Owner, "doSomething")
		require.Nil(t, err)

		ownerBalanceBefore := context.GetAccountBalance(&context.Owner).Uint64()
		reward := context.GetAccount(contractAddress).GetDeveloperReward().Uint64()

		err = context.ExecuteSC(&context.Owner, "ClaimDeveloperRewards")
		require.Nil(t, err)

		ownerBalanceAfter := context.GetAccountBalance(&context.Owner).Uint64()
		require.Equal(t, ownerBalanceBefore-context.LastConsumedFee+reward, ownerBalanceAfter)

		events := context.LastLogs[0].GetLogEvents()
		require.Equal(t, "ClaimDeveloperRewards", string(events[0].GetIdentifier()))
		require.Equal(t, big.NewInt(0).SetUint64(reward).Bytes(), events[0].GetTopics()[0])
		require.Equal(t, context.Owner.Address, events[0].GetTopics()[1])
	})

	t.Run("rewards for contract", func(t *testing.T) {
		parentContractAddress := context.ScAddress

		err = context.ExecuteSC(&context.Owner, "deployChild")
		require.Nil(t, err)

		chilContractdAddress := context.QuerySCBytes("getChildAddress", [][]byte{})
		require.NotNil(t, chilContractdAddress)

		context.ScAddress = chilContractdAddress
		err = context.ExecuteSC(&context.Owner, "doSomething")
		require.Nil(t, err)

		contractBalanceBefore := context.GetAccount(parentContractAddress).GetBalance().Uint64()
		reward := context.GetAccount(chilContractdAddress).GetDeveloperReward().Uint64()

		context.ScAddress = parentContractAddress
		err = context.ExecuteSC(&context.Owner, "claimDeveloperRewardsOnChild")
		require.Nil(t, err)

		contractBalanceAfter := context.GetAccount(parentContractAddress).GetBalance().Uint64()
		require.Equal(t, contractBalanceBefore+reward, contractBalanceAfter)

		events := context.LastLogs[0].GetLogEvents()
		require.Equal(t, "ClaimDeveloperRewards", string(events[0].GetIdentifier()))
		require.Equal(t, big.NewInt(0).SetUint64(reward).Bytes(), events[0].GetTopics()[0])
		require.Equal(t, parentContractAddress, events[0].GetTopics()[1])
	})
}
