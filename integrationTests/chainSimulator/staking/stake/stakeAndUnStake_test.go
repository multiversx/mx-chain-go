package stake

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/multiversx/mx-chain-core-go/core"
	coreAPI "github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/require"
)

const (
	defaultPathToInitialConfig = "../../../../cmd/node/config/"
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
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		NumNodesWaitingListMeta:     0,
		NumNodesWaitingListShard:    0,
		AlterConfigsFunction: func(cfg *config.Configs) {
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 8 // 8 nodes until new nodes will be placed on queue
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, numOfShards)
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
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(newValidatorOwnerBytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(newValidatorOwner, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeActiveValidator := accountValidatorOwner.Balance

	// Step 5 --- create an unStake transaction with the bls key of an initial validator and execute the transaction to make place for the validator that was added at step 3
	firstValidatorKey, err := cs.GetValidatorPrivateKeys()[0].GeneratePublic().ToByteArray()
	require.Nil(t, err)

	initialAddressWithValidators := cs.GetInitialWalletKeys().StakeWallets[0].Address
	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(initialAddressWithValidators.Bytes)
	initialAccount, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(initialAddressWithValidators.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	tx = &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   initialAddressWithValidators.Bytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(firstValidatorKey))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)

	// Step 6 --- generate 8 epochs to get rewards
	err = cs.GenerateBlocksUntilEpochIsReached(8)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = cs.ForceResetValidatorStatisticsCache()
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
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            100,
		MetaChainMinNodes:           100,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			cfg.GeneralConfig.ValidatorStatistics.CacheRefreshIntervalInSec = 1
			eligibleNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake
			// 8 nodes until new nodes will be placed on queue
			waitingNodes := uint32(8)
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(eligibleNodes), waitingNodes, numOfShards)
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

	txFromNetwork, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, txFromNetwork)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	err = cs.ForceResetValidatorStatisticsCache()
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
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            3,
		MetaChainMinNodes:           3,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriod = 1
			cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 1
			newNumNodes := cfg.SystemSCConfig.StakingSystemSCConfig.MaxNumberOfNodesForStake + 10
			configs.SetMaxNumberOfNodesInConfigs(cfg, uint32(newNumNodes), 0, numOfShards)
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

	mintValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	walletAddressShardID := uint32(0)
	walletAddress, err := cs.GenerateAndMintWalletAddress(walletAddressShardID, mintValue)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	bls0, _ := hex.DecodeString(blsKeys[0])
	blsKeyStatus := staking.GetBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "staked", blsKeyStatus)

	// do unStake
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, fmt.Sprintf("unStake@%s", blsKeys[0]), staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	blsKeyStatus = staking.GetBLSKeyStatus(t, metachainNode, bls0)
	require.Equal(t, "unStaked", blsKeyStatus)

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	// do unBond
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, fmt.Sprintf("unBondNodes@%s", blsKeys[0]), staking.GasLimitForStakeOperation)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	// do claim
	txClaim := chainSimulatorIntegrationTests.GenerateTransaction(walletAddress.Bytes, 3, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, "unBondTokens", staking.GasLimitForStakeOperation)
	claimTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txClaim, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, claimTx)

	err = cs.GenerateBlocks(5)
	require.Nil(t, err)

	// check tokens are in the wallet balance
	walletAccount, _, err := cs.GetNodeHandler(walletAddressShardID).GetFacadeHandler().GetAccount(walletAddress.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	walletBalanceBig, _ := big.NewInt(0).SetString(walletAccount.Balance, 10)
	require.True(t, walletBalanceBig.Cmp(chainSimulatorIntegrationTests.MinimumStakeValue) > 0)
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

// Test description
//  Stake funds - happy flow
//
//  Preconditions: have an account with egld and 2 staked nodes (2500 stake per node) - directly staked, and no unstake
//
//  1. Check the stake amount for the owner of the staked nodes with the vmquery "getTotalStaked", and the account current EGLD balance
//  2. Create from the owner of staked nodes a transaction to stake 1 EGLD and send it to the network
//  3. Check the outcome of the TX & verify new stake state with vmquery

// Internal test scenario #24
func TestChainSimulator_DirectStakingNodes_StakeFunds(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedNodesStakingFunds(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedNodesStakingFunds(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedNodesStakingFunds(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedNodesStakingFunds(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedNodesStakingFunds(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	log.Info("Preconditions. Have an account with 2 staked nodes")
	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(5010)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	stakeValue = big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	log.Info("Step 1. Check the stake amount for the owner of the staked nodes")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	log.Info("Step 2. Create from the owner of the staked nodes a tx to stake 1 EGLD")

	stakeValue = big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(1))
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	log.Info("Step 3. Check the stake amount for the owner of the staked nodes")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5001)
}

func checkExpectedStakedValue(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte, expectedValue int64) {
	totalStaked := getTotalStaked(t, metachainNode, blsKey)

	expectedStaked := big.NewInt(expectedValue)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(totalStaked))
}

func getTotalStaked(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) []byte {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	return result.ReturnData[0]
}

// Test description:
// Unstake funds with deactivation of node if below 2500 -> the rest of funds are distributed as topup at epoch change
//
// Internal test scenario #26
func TestChainSimulator_DirectStakingNodes_UnstakeFundsWithDeactivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	//  1. Check the stake amount and number of nodes for the owner of the staked nodes with the vmquery "getTotalStaked", and the account current EGLD balance
	//  2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network
	//  3. Check the outcome of the TX & verify new stake state with vmquery "getTotalStaked" and "getUnStakedTokensList"
	//  4. Wait for change of epoch and check the outcome

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedUnstakeFundsWithDeactivation(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(5010)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	stakeValue = big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[1])

	log.Info("Step 1. Check the stake amount for the owner of the staked nodes")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	log.Info("Step 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery getTotalStaked and getUnStakedTokensList")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 4990)

	unStakedTokensAmount := getUnStakedTokensList(t, metachainNode, validatorOwner.Bytes)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(unStakedTokensAmount).String())

	log.Info("Step 4. Wait for change of epoch and check the outcome")
	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	checkOneOfTheNodesIsUnstaked(t, metachainNode, blsKeys[:2])
}

func getUnStakedTokensList(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) []byte {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	return result.ReturnData[0]
}

func checkOneOfTheNodesIsUnstaked(t *testing.T,
	metachainNode chainSimulatorProcess.NodeHandler,
	blsKeys []string,
) {
	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	keyStatus0 := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	log.Info("Key info", "key", blsKeys[0], "status", keyStatus0)

	isNotStaked0 := keyStatus0 == staking.UnStakedStatus

	decodedBLSKey1, _ := hex.DecodeString(blsKeys[1])
	keyStatus1 := staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey1)
	log.Info("Key info", "key", blsKeys[1], "status", keyStatus1)

	isNotStaked1 := keyStatus1 == staking.UnStakedStatus

	require.True(t, isNotStaked0 != isNotStaked1)
}

func testBLSKeyStaked(t *testing.T,
	metachainNode chainSimulatorProcess.NodeHandler,
	blsKey string,
) {
	decodedBLSKey, _ := hex.DecodeString(blsKey)
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)

	validatorStatistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		require.Equal(t, staking.StakedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
		return
	}

	// in staking ph 2/3.5 we do not find the bls key on the validator statistics
	_, found := validatorStatistics[blsKey]
	require.False(t, found)
	require.Equal(t, staking.QueuedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
}

// Test description:
// Unstake funds with deactivation of node, followed by stake with sufficient ammount does not unstake node at end of epoch
//
// Internal test scenario #27
func TestChainSimulator_DirectStakingNodes_UnstakeFundsWithDeactivation_WithReactivation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	// 1. Check the stake amount and number of nodes for the owner of the staked nodes with the vmquery "getTotalStaked", and the account current EGLD balance
	// 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network
	// 3. Check the outcome of the TX & verify new stake state with vmquery
	// 4. Create from the owner of staked nodes a transaction to stake 10 EGLD and send it to the network
	// 5. Check the outcome of the TX & verify new stake state with vmquery
	// 6. Wait for change of epoch and check the outcome

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedUnstakeFundsWithDeactivationAndReactivation(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(6000)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	stakeValue = big.NewInt(0).Set(chainSimulatorIntegrationTests.MinimumStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[1])

	log.Info("Step 1. Check the stake amount for the owner of the staked nodes")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	log.Info("Step 2. Create from the owner of staked nodes a transaction to unstake 10 EGLD and send it to the network")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery getTotalStaked and getUnStakedTokensList")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 4990)

	unStakedTokensAmount := getUnStakedTokensList(t, metachainNode, validatorOwner.Bytes)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(unStakedTokensAmount).String())

	log.Info("Step 4. Create from the owner of staked nodes a transaction to stake 10 EGLD and send it to the network")

	newStakeValue := big.NewInt(10)
	newStakeValue = newStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, newStakeValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, newStakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	log.Info("5. Check the outcome of the TX & verify new stake state with vmquery")
	checkExpectedStakedValue(t, metachainNode, validatorOwner.Bytes, 5000)

	log.Info("Step 6. Wait for change of epoch and check the outcome")
	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])
	testBLSKeyStaked(t, metachainNode, blsKeys[1])
}

// Test description:
// Withdraw unstaked funds before unbonding period should return error
//
// Internal test scenario #28
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedFundsBeforeUnbonding(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	// 1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds
	// 2. Check the outcome of the TX & verify new stake state with vmquery ("getUnStakedTokensList")

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsBeforeUnbonding(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(10000)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	log.Info("Step 1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// check bls key is still staked
	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	log.Info("Step 2. Check the outcome of the TX & verify new stake state with vmquery (`getUnStakedTokensList`)")

	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	// the owner balance should decrease only with the txs fee
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	txsFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)
	balanceAfterUnbondingWithFee := big.NewInt(0).Add(balanceAfterUnbonding, txsFee)

	txsFee, _ = big.NewInt(0).SetString(unStakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	txsFee, _ = big.NewInt(0).SetString(stakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	require.Equal(t, 1, balanceAfterUnbondingWithFee.Cmp(balanceBeforeUnbonding))
}

// Test description:
// Withdraw unstaked funds in first available withdraw epoch
//
// Internal test scenario #29
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInWithdrawEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	// 1. Wait for the unbonding epoch to start
	// 2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds
	// 3. Check the outcome of the TX & verify new stake state with vmquery ("getUnStakedTokensList")

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInFirstEpoch(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(10000)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	unStakeValue := big.NewInt(10)
	unStakeValue = unStakeValue.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// check bls key is still staked
	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedUnStaked := big.NewInt(10)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	log.Info("Step 1. Wait for the unbonding epoch to start")

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	log.Info("Step 2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	log.Info("Step 3. Check the outcome of the TX & verify new stake state with vmquery (`getUnStakedTokensList`)")

	scQuery = &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err = metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedStaked := big.NewInt(2590)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(result.ReturnData[0]))

	// the owner balance should increase with the (10 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	// substract unbonding value
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue)

	txsFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)
	balanceAfterUnbondingWithFee := big.NewInt(0).Add(balanceAfterUnbonding, txsFee)

	txsFee, _ = big.NewInt(0).SetString(unStakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	txsFee, _ = big.NewInt(0).SetString(stakeTx.Fee, 10)
	balanceAfterUnbondingWithFee.Add(balanceAfterUnbondingWithFee, txsFee)

	require.Equal(t, 1, balanceAfterUnbondingWithFee.Cmp(balanceBeforeUnbonding))
}

// Test description:
// Unstaking funds in different batches allows correct withdrawal for each batch
// at the corresponding epoch.
//
// Internal test scenario #30
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInBatches(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	// 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.
	// 2. Send the transactions in consecutive epochs, one TX in each epoch.
	// 3. Wait for the epoch when first tx unbonding period ends.
	// 4. Create a transaction for withdraw and send it to the network
	// 5. Wait for an epoch
	// 6. Create another transaction for withdraw and send it to the network
	// 7. Wait for an epoch
	// 8. Create another transasction for withdraw and send it to the network

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 6
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInBatches(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(2700)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	stakeTxFee, _ := big.NewInt(0).SetString(stakeTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	log.Info("Step 1. Create 3 transactions for unstaking: first one unstaking 11 egld, second one unstaking 12 egld and third one unstaking 13 egld.")
	log.Info("Step 2. Send the transactions in consecutive epochs, one TX in each epoch.")

	unStakeValue1 := big.NewInt(11)
	unStakeValue1 = unStakeValue1.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue1)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue1.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	testEpoch := targetEpoch + 1
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	unStakeValue2 := big.NewInt(12)
	unStakeValue2 = unStakeValue2.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue2)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue2.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	unStakeValue3 := big.NewInt(13)
	unStakeValue3 = unStakeValue3.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue3)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue3.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	// check bls key is still staked
	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedUnStaked := big.NewInt(11)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	scQuery = &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err = metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedStaked := big.NewInt(2600 - 11 - 12 - 13)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(result.ReturnData[0]))

	log.Info("Step 3. Wait for the unbonding epoch to start")

	testEpoch += 3
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	log.Info("Step 4.1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 4, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	unBondTxFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)

	txsFee := big.NewInt(0)

	txsFee.Add(txsFee, stakeTxFee)
	txsFee.Add(txsFee, unBondTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)

	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))

	log.Info("Step 4.2. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 5, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ = big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)

	txsFee.Add(txsFee, unBondTxFee)
	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))

	log.Info("Step 4.3. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	testEpoch++
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 6, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12+13 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ = big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue3)

	txsFee.Add(txsFee, unBondTxFee)
	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))
}

// Test description:
// Unstake funds in different batches in the same epoch allows correct withdrawal in the correct epoch
//
// Internal test scenario #31
func TestChainSimulator_DirectStakingNodes_WithdrawUnstakedInEpoch(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test Steps
	// 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.
	// 2. Send the transactions consecutively in the same epoch
	// 3. Wait for the epoch when unbonding period ends.
	// 4. Create a transaction for withdraw and send it to the network

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:      false,
			TempDir:                     t.TempDir(),
			PathToInitialConfig:         defaultPathToInitialConfig,
			NumOfShards:                 3,
			GenesisTimestamp:            time.Now().Unix(),
			RoundDurationInMillis:       roundDurationInMillis,
			RoundsPerEpoch:              roundsPerEpoch,
			ApiInterface:                api.NewNoApiInterface(),
			MinNodesPerShard:            3,
			MetaChainMinNodes:           3,
			ConsensusGroupSize:          1,
			MetaChainConsensusGroupSize: 1,
			NumNodesWaitingListMeta:     3,
			NumNodesWaitingListShard:    3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4

				cfg.SystemSCConfig.StakingSystemSCConfig.UnBondPeriodInEpochs = 3
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t, cs, 4)
	})
}

func testChainSimulatorDirectStakedWithdrawUnstakedFundsInEpoch(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(2700)
	mintValue = mintValue.Mul(chainSimulatorIntegrationTests.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	stakeValue := big.NewInt(0).Mul(chainSimulatorIntegrationTests.OneEGLD, big.NewInt(2600))
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	stakeTxFee, _ := big.NewInt(0).SetString(stakeTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	shardIDValidatorOwner := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorOwner.Bytes)
	accountValidatorOwner, _, err := cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceBeforeUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	log.Info("Step 1. Create 3 transactions for unstaking: first one unstaking 11 egld each, second one unstaking 12 egld and third one unstaking 13 egld.")
	log.Info("Step 2. Send the transactions in consecutively in same epoch.")

	unStakeValue1 := big.NewInt(11)
	unStakeValue1 = unStakeValue1.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue1)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue1.Bytes()))
	txUnStake := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 1, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeTxFee, _ := big.NewInt(0).SetString(unStakeTx.Fee, 10)

	unStakeValue2 := big.NewInt(12)
	unStakeValue2 = unStakeValue2.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue2)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue2.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 2, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	unStakeValue3 := big.NewInt(13)
	unStakeValue3 = unStakeValue3.Mul(chainSimulatorIntegrationTests.OneEGLD, unStakeValue3)
	txDataField = fmt.Sprintf("unStakeTokens@%s", hex.EncodeToString(unStakeValue3.Bytes()))
	txUnStake = chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 3, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForStakeOperation)
	unStakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unStakeTx)

	// check bls key is still staked
	testBLSKeyStaked(t, metachainNode, blsKeys[0])

	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getUnStakedTokensList",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedUnStaked := big.NewInt(11 + 12 + 13)
	expectedUnStaked = expectedUnStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedUnStaked)
	require.Equal(t, expectedUnStaked.String(), big.NewInt(0).SetBytes(result.ReturnData[0]).String())

	scQuery = &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{validatorOwner.Bytes},
	}
	result, _, err = metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, chainSimulatorIntegrationTests.OkReturnCode, result.ReturnCode)

	expectedStaked := big.NewInt(2600 - 11 - 12 - 13)
	expectedStaked = expectedStaked.Mul(chainSimulatorIntegrationTests.OneEGLD, expectedStaked)
	require.Equal(t, expectedStaked.String(), string(result.ReturnData[0]))

	log.Info("Step 3. Wait for the unbonding epoch to start")

	testEpoch := targetEpoch + 3
	err = cs.GenerateBlocksUntilEpochIsReached(testEpoch)
	require.Nil(t, err)

	log.Info("Step 4.1. Create from the owner of staked nodes a transaction to withdraw the unstaked funds")

	txDataField = fmt.Sprintf("unBondTokens@%s", blsKeys[0])
	txUnBond := chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 4, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.ZeroValue, txDataField, staking.GasLimitForUnBond)
	unBondTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnBond, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unBondTx)

	unBondTxFee, _ := big.NewInt(0).SetString(unBondTx.Fee, 10)

	err = cs.GenerateBlocks(2)
	require.Nil(t, err)

	// the owner balance should increase with the (11+12+13 EGLD - tx fee)
	accountValidatorOwner, _, err = cs.GetNodeHandler(shardIDValidatorOwner).GetFacadeHandler().GetAccount(validatorOwner.Bech32, coreAPI.AccountQueryOptions{})
	require.Nil(t, err)
	balanceAfterUnbonding, _ := big.NewInt(0).SetString(accountValidatorOwner.Balance, 10)

	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue1)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue2)
	balanceAfterUnbonding.Sub(balanceAfterUnbonding, unStakeValue3)

	txsFee := big.NewInt(0)

	txsFee.Add(txsFee, stakeTxFee)
	txsFee.Add(txsFee, unBondTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)
	txsFee.Add(txsFee, unStakeTxFee)

	balanceAfterUnbonding.Add(balanceAfterUnbonding, txsFee)

	require.Equal(t, 1, balanceAfterUnbonding.Cmp(balanceBeforeUnbonding))
}

// Test that if we unStake one active node(waiting/eligible), the number of qualified nodes will remain the same
// Nodes configuration at genesis consisting of a total of 32 nodes, distributed on 3 shards + meta:
// - 4 eligible nodes/shard
// - 4 waiting nodes/shard
// - 2 nodes to shuffle per shard
// - max num nodes config for stakingV4 step3 = 24 (being downsized from previously 32 nodes)
// - with this config, we should always select 8 nodes from auction list
// We will add one extra node, so auction list size = 9, but will always select 8. Even if we unStake one active node,
// we should still only select 8 nodes.
func TestChainSimulator_UnStakeOneActiveNodeAndCheckAPIAuctionList(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	stakingV4Step1Epoch := uint32(2)
	stakingV4Step2Epoch := uint32(3)
	stakingV4Step3Epoch := uint32(4)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            4,
		MetaChainMinNodes:           4,
		ConsensusGroupSize:          1,
		MetaChainConsensusGroupSize: 1,
		NumNodesWaitingListMeta:     4,
		NumNodesWaitingListShard:    4,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = stakingV4Step1Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = stakingV4Step2Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.CleanupAuctionOnLowWaitingListEnableEpoch = stakingV4Step1Epoch

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].MaxNumNodes = 32
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 24
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 2
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	err = cs.GenerateBlocksUntilEpochIsReached(int32(stakingV4Step3Epoch + 1))
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	qualified, unQualified := getQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 0, len(unQualified))

	stakeOneNode(t, cs)

	qualified, unQualified = getQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 1, len(unQualified))

	unStakeOneActiveNode(t, cs)

	qualified, unQualified = getQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 8, len(qualified))
	require.Equal(t, 1, len(unQualified))
}

// Nodes configuration at genesis consisting of a total of 40 nodes, distributed on 3 shards + meta:
// - 4 eligible nodes/shard
// - 4 waiting nodes/shard
// - 2 nodes to shuffle per shard
// - max num nodes config for stakingV4 step3 = 32 (being downsized from previously 40 nodes)
// - with this config, we should always select max 8 nodes from auction list if there are > 40 nodes in the network
// This test will run with only 32 nodes and check that there are no nodes in the auction list,
// because of the lowWaitingList condition being triggered when in staking v4
func TestChainSimulator_EdgeCaseLowWaitingList(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	startTime := time.Now().Unix()
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	stakingV4Step1Epoch := uint32(2)
	stakingV4Step2Epoch := uint32(3)
	stakingV4Step3Epoch := uint32(4)

	numOfShards := uint32(3)
	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck:      false,
		TempDir:                     t.TempDir(),
		PathToInitialConfig:         defaultPathToInitialConfig,
		NumOfShards:                 numOfShards,
		GenesisTimestamp:            startTime,
		RoundDurationInMillis:       roundDurationInMillis,
		RoundsPerEpoch:              roundsPerEpoch,
		ApiInterface:                api.NewNoApiInterface(),
		MinNodesPerShard:            4,
		MetaChainMinNodes:           4,
		NumNodesWaitingListMeta:     2,
		NumNodesWaitingListShard:    2,
		MetaChainConsensusGroupSize: 1,
		ConsensusGroupSize:          1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = stakingV4Step1Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = stakingV4Step2Epoch
			cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = stakingV4Step3Epoch

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].MaxNumNodes = 40
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[1].NodesToShufflePerShard = 2

			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = stakingV4Step3Epoch
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].MaxNumNodes = 32
			cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].NodesToShufflePerShard = 2
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)

	defer cs.Close()

	epochToCheck := int32(stakingV4Step3Epoch + 1)
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	qualified, unQualified := getQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 0, len(qualified))
	require.Equal(t, 0, len(unQualified))

	// we always have 0 in auction list because of the lowWaitingList condition
	epochToCheck += 1
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	qualified, unQualified = getQualifiedAndUnqualifiedNodes(t, metachainNode)
	require.Equal(t, 0, len(qualified))
	require.Equal(t, 0, len(unQualified))

	// stake 16 mode nodes, these will go to auction list
	stakeNodes(t, cs, 17)

	epochToCheck += 1
	err = cs.GenerateBlocksUntilEpochIsReached(epochToCheck)
	require.Nil(t, err)

	qualified, unQualified = getQualifiedAndUnqualifiedNodes(t, metachainNode)
	// all the previously registered will be selected, as we have 24 nodes in eligible+waiting, 8 will shuffle out,
	// but this time there will be not be lowWaitingList, as there are enough in auction, so we will end up with
	// 24-8 = 16 nodes remaining + 16 from auction, to fill up all 32 positions
	require.Equal(t, 16, len(qualified))
	require.Equal(t, 1, len(unQualified))

	shuffledOutNodesKeys, err := metachainNode.GetProcessComponents().NodesCoordinator().GetShuffledOutToAuctionValidatorsPublicKeys(uint32(epochToCheck))
	require.Nil(t, err)

	checkKeysNotInMap(t, shuffledOutNodesKeys, qualified)
	checkKeysNotInMap(t, shuffledOutNodesKeys, unQualified)
}

func checkKeysNotInMap(t *testing.T, m map[uint32][][]byte, keys []string) {
	for _, key := range keys {
		for _, v := range m {
			for _, k := range v {
				mapKey := hex.EncodeToString(k)
				require.NotEqual(t, key, mapKey)
			}
		}
	}
}

func stakeNodes(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, numNodesToStake int) {
	txs := make([]*transaction.Transaction, numNodesToStake)
	for i := 0; i < numNodesToStake; i++ {
		txs[i] = createStakeTransaction(t, cs)
	}

	stakeTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted(txs, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTxs)
	require.Len(t, stakeTxs, numNodesToStake)

	require.Nil(t, cs.GenerateBlocks(1))
}

func stakeOneNode(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) {
	txStake := createStakeTransaction(t, cs)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	require.Nil(t, cs.GenerateBlocks(1))
}

func createStakeTransaction(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) *transaction.Transaction {
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)

	mintValue := big.NewInt(0).Add(chainSimulatorIntegrationTests.MinimumStakeValue, chainSimulatorIntegrationTests.OneEGLD)
	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	return chainSimulatorIntegrationTests.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, chainSimulatorIntegrationTests.MinimumStakeValue, txDataField, staking.GasLimitForStakeOperation)
}

func unStakeOneActiveNode(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator) {
	err := cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	validators, err := cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	idx := 0
	keyToUnStake := make([]byte, 0)
	numKeys := len(cs.GetValidatorPrivateKeys())
	for idx = 0; idx < numKeys; idx++ {
		keyToUnStake, err = cs.GetValidatorPrivateKeys()[idx].GeneratePublic().ToByteArray()
		require.Nil(t, err)

		apiValidator, found := validators[hex.EncodeToString(keyToUnStake)]
		require.True(t, found)

		validatorStatus := apiValidator.ValidatorStatus
		if validatorStatus == "waiting" || validatorStatus == "eligible" {
			log.Info("found active key to unStake", "index", idx, "bls key", keyToUnStake, "list", validatorStatus)
			break
		}

		if idx == numKeys-1 {
			require.Fail(t, "did not find key to unStake")
		}
	}

	rcv := "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l"
	rcvAddrBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(rcv)

	validatorWallet := cs.GetInitialWalletKeys().StakeWallets[idx].Address
	shardID := cs.GetNodeHandler(0).GetShardCoordinator().ComputeId(validatorWallet.Bytes)
	initialAccount, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().GetAccount(validatorWallet.Bech32, coreAPI.AccountQueryOptions{})

	require.Nil(t, err)
	tx := &transaction.Transaction{
		Nonce:     initialAccount.Nonce,
		Value:     big.NewInt(0),
		SndAddr:   validatorWallet.Bytes,
		RcvAddr:   rcvAddrBytes,
		Data:      []byte(fmt.Sprintf("unStake@%s", hex.EncodeToString(keyToUnStake))),
		GasLimit:  50_000_000,
		GasPrice:  1000000000,
		Signature: []byte("dummy"),
		ChainID:   []byte(configs.ChainID),
		Version:   1,
	}
	_, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	err = cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)
	validators, err = cs.GetNodeHandler(core.MetachainShardId).GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)

	apiValidator, found := validators[hex.EncodeToString(keyToUnStake)]
	require.True(t, found)
	require.True(t, strings.Contains(apiValidator.ValidatorStatus, "leaving"))
}
