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
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig             = "../../../cmd/node/config/"
	maxNumOfBlockToGenerateWhenExecutingTx = 7
)

var log = logger.GetOrCreate("integrationTests/chainSimulator")

// Test scenario
// 1. Add a new validator private key in the multi key handler
// 2. Do a stake transaction for the validator key
// 3. Do an unstake transaction (to make a place for the new validator)
// 4. Check if the new validator has generated rewards
func TestChainSimulator_AddValidatorKey(t *testing.T) {
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
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
	})
	require.Nil(t, err)
	require.NotNil(t, cm)

	err = cm.GenerateBlocks(30)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	privateKeyBase64 := "NjRhYjk3NmJjYWVjZTBjNWQ4YmJhNGU1NjZkY2VmYWFiYjcxNDI1Y2JiZDcwYzc1ODA2MGUxNTE5MGM2ZjE1Zg=="
	helpers.AddValidatorKeysInMultiKey(t, cm, []string{privateKeyBase64})

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	blsKey := "9b7de1b2d2c90b7bea8f6855075c77d6c63b5dada29abb9b87c52cfae9d4112fcac13279e1a07d94672a5e62a83e3716555513014324d5c6bb4261b465f1b8549a7a338bc3ae8edc1e940958f9c2e296bd3c118a4466dec99dda0ceee3eb6a8c"

	// Step 3 --- generate and send a stake transaction with the BLS key of the validator key that was added at step 1
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@01@%s@010101", blsKey)),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_ = helpers.SendTxAndGenerateBlockTilTxIsExecuted(t, cm, tx, maxNumOfBlockToGenerateWhenExecutingTx)

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
	tx = &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   senderBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(firstValidatorKey))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_ = helpers.SendTxAndGenerateBlockTilTxIsExecuted(t, cm, tx, maxNumOfBlockToGenerateWhenExecutingTx)

	// Step 6 --- generate 50 blocks to pass 2 epochs and the validator to generate rewards
	err = cm.GenerateBlocks(50)
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

func TestChainSimulator_AddANewValidatorAfterStakingV4(t *testing.T) {
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
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
	})
	require.Nil(t, err)
	require.NotNil(t, cm)

	err = cm.GenerateBlocks(150)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	numOfNodes := 10
	validatorSecretKeysBytes, blsKeys := helpers.GenerateBlsPrivateKeys(t, numOfNodes)
	err = cm.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cm.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cm.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "100000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS key of the validator key that was added at step 1
	validatorData := ""
	for _, blsKey := range blsKeys {
		validatorData += fmt.Sprintf("@%s@010101", blsKey)
	}

	log.Warn("BLS KEYS", "keys", validatorData)

	numOfNodesHex := hex.EncodeToString(big.NewInt(int64(numOfNodes)).Bytes())
	stakeValue, _ := big.NewInt(0).SetString("25000000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@%s%s", numOfNodesHex, validatorData)),
		GasLimit:  500_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}

	_ = logger.SetLogLevel("*:DEBUG")

	txFromNetwork := helpers.SendTxAndGenerateBlockTilTxIsExecuted(t, cm, tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.NotNil(t, txFromNetwork)

	err = cm.GenerateBlocks(20)
	require.Nil(t, err)
}
