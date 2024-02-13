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
	"github.com/multiversx/mx-chain-core-go/data/validator"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig             = "../../../cmd/node/config/"
	maxNumOfBlockToGenerateWhenExecutingTx = 7
)

var log = logger.GetOrCreate("integrationTests/chainSimulator")

// TODO scenarios
// Make a staking provider with max num of nodes
// DO a merge transaction

// Test scenario
// 1. Add a new validator private key in the multi key handler
// 2. Do a stake transaction for the validator key
// 3. Do an unstake transaction (to make a place for the new validator)
// 4. Check if the new validator has generated rewards
func TestChainSimulator_AddValidatorKey(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:   false,
		TempDir:                  t.TempDir(),
		PathToInitialConfig:      defaultPathToInitialConfig,
		NumOfShards:              numOfShards,
		GenesisTimestamp:         startTime,
		RoundDurationInMillis:    roundDurationInMillis,
		RoundsPerEpoch:           roundsPerEpoch,
		ApiInterface:             api.NewNoApiInterface(),
		MinNodesPerShard:         3,
		MetaChainMinNodes:        3,
		NumNodesWaitingListMeta:  0,
		NumNodesWaitingListShard: 0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(30)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "10000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS key of the validator key that was added at step 1
	stakeValue, _ := big.NewInt(0).SetString("2500000000000000000000", 10)
	tx := &transaction.Transaction{
		Nonce:     0,
		Value:     stakeValue,
		SndAddr:   newValidatorOwnerBytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("stake@01@%s@010101", blsKeys[0])),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(newValidatorOwnerBytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeActiveValidator := accountValidatorOwner.Balance

	// Step 5 --- create an unStake transaction with the bls key of an initial validator and execute the transaction to make place for the validator that was added at step 3
	firstValidatorKey, err := cs.GetValidatorPrivateKeys()[0].GeneratePublic().ToByteArray()
	require.Nil(t, err)

	initialAddressWithValidators := cs.GetInitialWalletKeys().InitialWalletWithStake.Address
	senderBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(initialAddressWithValidators)
	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(senderBytes)
	initialAccount, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(initialAddressWithValidators, coreAPI.AccountQueryOptions{})
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
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)

	// Step 6 --- generate 8 epochs to get rewards
	err = cs.GenerateBlocksUntilEpochIsReached(8)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	validatorStatistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	checkValidatorsRating(t, validatorStatistics)

	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
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
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}
	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       100,
		MetaChainMinNodes:      100,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.GeneralConfig.ValidatorStatistics.CacheRefreshIntervalInSec = 1
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocks(150)
	require.Nil(t, err)

	// Step 1 --- add a new validator key in the chain simulator
	numOfNodes := 20
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(numOfNodes)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	newValidatorOwner := "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl"
	newValidatorOwnerBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(newValidatorOwner)
	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	// Step 2 --- set an initial balance for the address that will initialize all the transactions
	err = cs.SetStateMultiple([]*dtos.AddressState{
		{
			Address: "erd1l6xt0rqlyzw56a3k8xwwshq2dcjwy3q9cppucvqsmdyw8r98dz3sae0kxl",
			Balance: "1000000000000000000000000",
		},
	})
	require.Nil(t, err)

	// Step 3 --- generate and send a stake transaction with the BLS keys of the validators key that were added at step 1
	validatorData := ""
	for _, blsKey := range blsKeys {
		validatorData += fmt.Sprintf("@%s@010101", blsKey)
	}

	numOfNodesHex := hex.EncodeToString(big.NewInt(int64(numOfNodes)).Bytes())
	stakeValue, _ := big.NewInt(0).SetString("51000000000000000000000", 10)
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

	txFromNetwork, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	results, err := metachainNode.GetFacadeHandler().AuctionListApi()
	require.Nil(t, err)
	require.Equal(t, newValidatorOwner, results[0].Owner)
	require.Equal(t, 20, len(results[0].Nodes))
	checkTotalQualified(t, results, 8)

	err = cs.GenerateBlocks(100)
	require.Nil(t, err)

	results, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().AuctionListApi()
	require.Nil(t, err)
	checkTotalQualified(t, results, 0)
}

// Internal test scenario #4 #5 #6
// do stake
// do unStake
// do unBondNodes
// do unBondTokens
func TestChainSimulatorStakeUnStakeUnBond(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 1)
	})

	t.Run("staking ph 4 step 1 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 4)
	})

	t.Run("staking ph 4 step 2 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 5)
	})

	t.Run("staking ph 4 step 3 active", func(t *testing.T) {
		testStakeUnStakeUnBond(t, 6)
	})
}

func testStakeUnStakeUnBond(t *testing.T, targetEpoch int32) {
	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}
	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: false,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            numOfShards,
		GenesisTimestamp:       startTime,
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       3,
		MetaChainMinNodes:      3,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod = 1
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 10
			configs.SetMaxNumberOfNodesInConfigs(cfg, newNumNodes, numOfShards)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Mul(oneEGLD, big.NewInt(2600))
	walletAddressShardID := uint32(0)
	walletAddress, err := cs.GenerateAndMintWalletAddress(walletAddressShardID, mintValue)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], mockBLSSignature)
	txStake := generateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, minimumStakeValue, txDataField, gasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	bls0, _ := hex.DecodeString(blsKeys[0])
	blsKeyStatus := getBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "staked", blsKeyStatus)

	// do unStake
	txUnStake := generateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, zeroValue, fmt.Sprintf("unStake@%s", blsKeys[0]), gasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	blsKeyStatus = getBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "unStaked", blsKeyStatus)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	// do unBond
	txUnBond := generateTransaction(walletAddress.Bytes, 2, vm.ValidatorSCAddress, zeroValue, fmt.Sprintf("unBondNodes@%s", blsKeys[0]), gasLimitForStakeOperation)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	// do claim
	txClaim := generateTransaction(walletAddress.Bytes, 3, vm.ValidatorSCAddress, zeroValue, "unBondTokens", gasLimitForStakeOperation)
	claimTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txClaim, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, claimTx)

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// check tokens are in the wallet balance
	walletAccount, _, err := cs.GetNodeHandler(walletAddressShardID).GetFacadeHandler().GetAccount(walletAddress.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	walletBalanceBig, _ := big.NewInt(0).SetString(walletAccount.Balance, 10)
	require.True(t, walletBalanceBig.Cmp(minimumStakeValue) > 0)
}

func checkTotalQualified(t *testing.T, auctionList []*common.AuctionListValidatorAPIResponse, expected int) {
	totalQualified := 0
	for _, res := range auctionList {
		for _, node := range res.Nodes {
			if node.Qualified {
				totalQualified++
			}
		}
	}
	require.Equal(t, expected, totalQualified)
}

func checkValidatorsRating(t *testing.T, validatorStatistics map[string]*validator.ValidatorStatistics) {
	countRatingIncreased := 0
	for _, validatorInfo := range validatorStatistics {
		validatorSignedAtLeastOneBlock := validatorInfo.NumValidatorSuccess > 0 || validatorInfo.NumLeaderSuccess > 0
		if !validatorSignedAtLeastOneBlock {
			continue
		}
		countRatingIncreased++
		require.Greater(t, validatorInfo.TempRating, validatorInfo.Rating)
	}
	require.Greater(t, countRatingIncreased, 0)
}
