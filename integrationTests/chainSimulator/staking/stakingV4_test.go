package staking

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/helpers"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/stretchr/testify/require"
)

// TODO scenarios
// Make a staking provider with max num of nodes
// DO a merge transaction

// Test scenario
// 1. Add a new validator private key in the multi key handler
// 2. Do a stake transaction for the validator key
// 3. Do an unstake transaction (to make a place for the new validator)
// 4. Check if the new validator has generated rewards
func TestChainSimulator_Initial_Setup(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip("this is not a short test")
	// }

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cm, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       6,
		MetaChainMinNodes:      6,
	})
	require.Nil(t, err)
	require.NotNil(t, cm)

	err = cm.GenerateBlocks(30)
	require.Nil(t, err)

	// Step 1 --- three new validator keys in the chain simulator
	privateKeyBase64_1 := "NjRhYjk3NmJjYWVjZTBjNWQ4YmJhNGU1NjZkY2VmYWFiYjcxNDI1Y2JiZDcwYzc1ODA2MGUxNTE5MGM2ZjE1Zg=="
	privateKeyBase64_2 := "NmVjYTAwNzczYjUwMjUyNmE0YzhlN2VjYTFlOTZlMzIyZmU3ODk5NWM2MzYyY2U0ZDQyYmRlYjI1YjgyZGE0NA=="
	privateKeyBase64_3 := "NWM5YWVkNWRmMGM0NjdkMTRlOTQ2OWMxNWRjNDliOGM4OWMxNGNiNzM4NGM1M2I0MjI2NDExODIxNTRmNTA2ZQ=="
	helpers.AddValidatorKeysInMultiKey(t, cm, []string{privateKeyBase64_1, privateKeyBase64_2, privateKeyBase64_3})

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	stakingContractAddr := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	stakingContractAddrAddrBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(stakingContractAddr)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions - 100_000 EGLD
	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "100000000000000000000000",
		},
	})
	require.Nil(t, err)

	//add the three blskeys
	blsKeys := []string{
		"9b7de1b2d2c90b7bea8f6855075c77d6c63b5dada29abb9b87c52cfae9d4112fcac13279e1a07d94672a5e62a83e3716555513014324d5c6bb4261b465f1b8549a7a338bc3ae8edc1e940958f9c2e296bd3c118a4466dec99dda0ceee3eb6a8c",
		"393dffaea5e356963b38d85e3468bf6ba30d4ebf26c7f1f7bc5cd0448741b64605ecbc07ca145d1adfefb828677b58052ce14ab173bc33ee874e427be84d25a5e807805eb10bb7aede3e4716339a8fc5086cbacfea87232dbe6643a963bd5d8d",
		"204ee6c9a68a6a0d5a4af5426d5a34b1ff0a5c62d0f93ee09aabbc54ad7f864b1f7c69d6d832d98425ce8626884cb611248e1004c02bb95acb821231195f388c8cc2dd7cb79c4f35d77bfb814dc5e5e297013252622320374eb744beb26a4794",
	}

	var nonce uint64 = 0
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	transactionsMap := make(map[string]*transaction.Transaction)
	for _, blsKey := range blsKeys {
		tx := &transaction.Transaction{
			Nonce:     nonce,
			Value:     stakeValue,
			SndAddr:   newValidatorOwnerBytes,
			RcvAddr:   stakingContractAddrAddrBytes,
			Data:      []byte(fmt.Sprintf("stake@01@%s@010101", blsKey)),
			GasLimit:  50_000_000,
			GasPrice:  1000000000,
			Signature: []byte("dummy"),
			ChainID:   []byte(configs.ChainID),
			Version:   1,
		}
		_ = helpers.SendTxAndGenerateBlockTilTxIsExecuted(t, cm, tx, maxNumOfBlockToGenerateWhenExecutingTx)
		nonce++

		transactionsMap[blsKey] = tx
	}

	shardIDValidatorOwner := cm.GetNodeHandler(0).GetShardCoordinator().ComputeId(newValidatorOwnerBytes)
	accountValidatorOwner, _, err := cm.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeActiveValidator := accountValidatorOwner.Balance

	// Step 5 --- create an unStake transaction with the bls key of an initial validator and execute the transaction to make place for the validator that was added at step 3
	firstValidatorKey, err := cm.GetValidatorPrivateKeys()[0].GeneratePublic().ToByteArray()
	require.Nil(t, err)

	initialAddressWithValidators := cm.GetInitialWalletKeys().InitialWalletWithStake.Address
	senderBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddressWithValidators)
	shardID := cm.GetNodeHandler(0).GetShardCoordinator().ComputeId(senderBytes)
	initialAccount, _, err := cm.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(initialAddressWithValidators, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	unstakeTx := &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   senderBytes,
		RcvAddr:   stakingContractAddrAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(firstValidatorKey))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_ = helpers.SendTxAndGenerateBlockTilTxIsExecuted(t, cm, unstakeTx, maxNumOfBlockToGenerateWhenExecutingTx)

	// Step 6 --- generate 100 blocks to pass 4 epochs and the validator to generate rewards
	err = cm.GenerateBlocks(80)
	require.Nil(t, err)

	accountValidatorOwner, _, err = cm.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterActiveValidator := accountValidatorOwner.Balance

	log.Info("balance before validator", "value", balanceBeforeActiveValidator)
	log.Info("balance after validator", "value", balanceAfterActiveValidator)

	balanceBeforeBig, _ := big.NewInt(0).SetString(balanceBeforeActiveValidator, 10)
	balanceAfterBig, _ := big.NewInt(0).SetString(balanceAfterActiveValidator, 10)
	diff := balanceAfterBig.Sub(balanceAfterBig, balanceBeforeBig)
	log.Info("difference", "value", diff.String())

	// Step 7 --- check the balance of the validator owner has been increased
	require.True(t, diff.Cmp(big.NewInt(0)) > 0)
}
