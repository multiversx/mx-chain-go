package stakingProvider

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-core-go/data/validator"
	dataVm "github.com/multiversx/mx-chain-core-go/data/vm"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/dtos"
	chainSimulatorProcess "github.com/multiversx/mx-chain-go/node/chainSimulator/process"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/vm"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var log = logger.GetOrCreate("stakingProvider")

const gasLimitForConvertOperation = 510_000_000
const gasLimitForDelegationContractCreationOperation = 500_000_000
const gasLimitForAddNodesOperation = 500_000_000
const gasLimitForUndelegateOperation = 500_000_000
const gasLimitForMergeOperation = 600_000_000
const gasLimitForDelegate = 12_000_000

const maxCap = "00"          // no cap
const hexServiceFee = "0ea1" // 37.45%
const walletAddressBytesLen = 32

// Test description:
//  Test that delegation contract created with MakeNewContractFromValidatorData works properly
//  Also check that delegate and undelegate works properly and the top-up remain the same if every delegator undelegates.
//  Test that the top-up from normal stake will be transferred after creating the contract and will be used in auction list computing

// Internal test scenario #10
func TestChainSimulator_MakeNewContractFromValidatorData(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on queue and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				maxNodesChangeEnableEpoch := cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch
				blsMultiSignerEnableEpoch := cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch

				cfg.EpochConfig.EnableEpochs = config.EnableEpochs{}
				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = maxNodesChangeEnableEpoch
				cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch = blsMultiSignerEnableEpoch

				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 1)
	})

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on queue and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 is not active and all is done in epoch 0", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				maxNodesChangeEnableEpoch := cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch
				blsMultiSignerEnableEpoch := cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch

				// set all activation epoch values on 0
				cfg.EpochConfig.EnableEpochs = config.EnableEpochs{}
				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch = maxNodesChangeEnableEpoch
				cfg.EpochConfig.EnableEpochs.BLSMultiSignerEnableEpoch = blsMultiSignerEnableEpoch

				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 100
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 101
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 102

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 102
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		// we need a little time to enable the VM queries on the http server
		time.Sleep(time.Second)
		// also, propose a couple of blocks
		err = cs.GenerateBlocks(3)
		require.Nil(t, err)

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 0)
	})

	// Test scenario done in staking v4 phase step 1
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 2)
	})

	// Test scenario done in staking v4 phase step 2
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 3)
	})

	// Test scenario done in staking v4 phase step 3
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Do a stake transaction for the validator key and test that the new key is on auction list and topup is 500
	// 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on auction list and topup is 500
	// 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700
	// 6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorData(t, cs, 4)
	})
}

func testChainSimulatorMakeNewContractFromValidatorData(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	log.Info("Step 1. Add a new validator private key in the multi key handler")
	privateKey, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKey)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	log.Info("Step 2. Set the initial state for the owner and the 2 delegators")
	mintValue := big.NewInt(3010)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	validatorOwner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	delegator1, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	delegator2, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	log.Info("working with the following addresses",
		"newValidatorOwner", validatorOwner.Bech32, "delegator1", delegator1.Bech32, "delegator2", delegator2.Bech32)

	log.Info("Step 3. Do a stake transaction for the validator key and test that the new key is on queue / auction list and the correct topup")
	stakeValue := big.NewInt(0).Set(staking.MinimumStakeValue)
	addedStakedValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(500))
	stakeValue.Add(stakeValue, addedStakedValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := staking.GenerateTransaction(validatorOwner.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorOwner.Bytes, blsKeys[0], addedStakedValue, 1)

	log.Info("Step 4. Execute the MakeNewContractFromValidatorData transaction and test that the key is on queue / auction list and the correct topup")
	txDataField = fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, hexServiceFee)
	txConvert := staking.GenerateTransaction(validatorOwner.Bytes, 1, vm.DelegationManagerSCAddress, staking.ZeroValue, txDataField, gasLimitForConvertOperation)
	convertTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	delegationAddress := convertTx.Logs.Events[0].Topics[1]
	delegationAddressBech32 := metachainNode.GetCoreComponents().AddressPubKeyConverter().SilentEncode(delegationAddress, log)
	log.Info("generated delegation address", "address", delegationAddressBech32)

	err = cs.ForceResetValidatorStatisticsCache()
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], addedStakedValue, 1)

	log.Info("Step 5. Execute 2 delegation operations of 100 EGLD each, check the topup is 700")
	delegateValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(100))
	txDelegate1 := staking.GenerateTransaction(delegator1.Bytes, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	txDelegate2 := staking.GenerateTransaction(delegator2.Bytes, 0, delegationAddress, delegateValue, "delegate", gasLimitForDelegate)
	delegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate2, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate2Tx)

	expectedTopUp := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(700))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], expectedTopUp, 1)

	log.Info("6. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 500")
	unDelegateValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(100))
	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate1 := staking.GenerateTransaction(delegator1.Bytes, 1, delegationAddress, staking.ZeroValue, txDataField, gasLimitForDelegate)
	unDelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate1Tx)

	txDataField = fmt.Sprintf("unDelegate@%s", hex.EncodeToString(unDelegateValue.Bytes()))
	txUnDelegate2 := staking.GenerateTransaction(delegator2.Bytes, 1, delegationAddress, staking.ZeroValue, txDataField, gasLimitForDelegate)
	unDelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUnDelegate2, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, unDelegate2Tx)

	expectedTopUp = big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(500))
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], expectedTopUp, 1)
}

func testBLSKeyIsInQueueOrAuction(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte, blsKey string, expectedTopUp *big.Int, actionListSize int) {
	decodedBLSKey, _ := hex.DecodeString(blsKey)
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	assert.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, address))

	activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		testBLSKeyIsInAuction(t, metachainNode, decodedBLSKey, blsKey, expectedTopUp, actionListSize, statistics, 1, address)
		return
	}

	// in staking ph 2/3.5 we do not find the bls key on the validator statistics
	_, found := statistics[blsKey]
	require.False(t, found)
	require.Equal(t, staking.QueuedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
}

func testBLSKeyIsInAuction(
	t *testing.T,
	metachainNode chainSimulatorProcess.NodeHandler,
	blsKeyBytes []byte,
	blsKey string,
	topUpInAuctionList *big.Int,
	actionListSize int,
	validatorStatistics map[string]*validator.ValidatorStatistics,
	numNodes int,
	owner []byte,
) {
	require.Equal(t, staking.StakedStatus, staking.GetBLSKeyStatus(t, metachainNode, blsKeyBytes))

	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	currentEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch()
	if metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step2Flag) == currentEpoch {
		// starting from phase 2, we have the shuffled out nodes from the previous epoch in the action list
		actionListSize += 8
	}
	if metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step3Flag) <= currentEpoch {
		// starting from phase 3, we have the shuffled out nodes from the previous epoch in the action list
		actionListSize += 4
	}

	require.Equal(t, actionListSize, len(auctionList))
	ownerAsBech32, err := metachainNode.GetCoreComponents().AddressPubKeyConverter().Encode(owner)
	require.Nil(t, err)
	if actionListSize != 0 {
		nodeWasFound := false
		for _, item := range auctionList {
			if item.Owner != ownerAsBech32 {
				continue
			}

			require.Equal(t, numNodes, len(auctionList[0].Nodes))
			for _, node := range item.Nodes {
				if node.BlsKey == blsKey {
					require.Equal(t, topUpInAuctionList.String(), item.TopUpPerNode)
					nodeWasFound = true
				}
			}
		}
		require.True(t, nodeWasFound)
	}

	// in staking ph 4 we should find the key in the validators statics
	validatorInfo, found := validatorStatistics[blsKey]
	require.True(t, found)
	require.Equal(t, staking.AuctionStatus, validatorInfo.ValidatorStatus)
}

func testBLSKeysAreInQueueOrAuction(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte, blsKeys []string, totalTopUp *big.Int, actionListSize int) {
	err := metachainNode.GetProcessComponents().ValidatorsProvider().ForceUpdate()
	require.Nil(t, err)
	statistics, err := metachainNode.GetFacadeHandler().ValidatorStatisticsApi()
	require.Nil(t, err)
	assert.Equal(t, totalTopUp, getBLSTopUpValue(t, metachainNode, address))

	individualTopup := big.NewInt(0).Set(totalTopUp)
	individualTopup.Div(individualTopup, big.NewInt(int64(len(blsKeys))))

	for _, blsKey := range blsKeys {
		decodedBLSKey, _ := hex.DecodeString(blsKey)
		activationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
		if activationEpoch <= metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
			testBLSKeyIsInAuction(t, metachainNode, decodedBLSKey, blsKey, individualTopup, actionListSize, statistics, len(blsKeys), address)
			continue
		}

		// in staking ph 2/3.5 we do not find the bls key on the validator statistics
		_, found := statistics[blsKey]
		require.False(t, found)
		require.Equal(t, staking.QueuedStatus, staking.GetBLSKeyStatus(t, metachainNode, decodedBLSKey))
	}
}

// Test description:
// Test that 2 different contracts with different topups that came from the normal stake will be considered in auction list computing in the correct order
// 1. Add 2 new validator private keys in the multi key handler
// 2. Set the initial state for 2 owners (mint 2 new wallets)
// 3. Do 2 stake transactions and test that the new keys are on queue / auction list and have the correct topup - 100 and 200 EGLD, respectively
// 4. Convert both validators into staking providers and test that the new keys are on queue / auction list and have the correct topup
// 5. If the staking v4 is activated (regardless the steps), check that the auction list sorted the 2 BLS keys based on topup

// Internal test scenario #11
func TestChainSimulator_MakeNewContractFromValidatorDataWith2StakingContracts(t *testing.T) {
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
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorDataWith2StakingContracts(t, cs, 1)
	})
	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorDataWith2StakingContracts(t, cs, 2)
	})
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorDataWith2StakingContracts(t, cs, 3)
	})
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMakeNewContractFromValidatorDataWith2StakingContracts(t, cs, 4)
	})
}

func testChainSimulatorMakeNewContractFromValidatorDataWith2StakingContracts(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	log.Info("Step 1. Add 2 new validator private keys in the multi key handler")
	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(2)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	log.Info("Step 2. Set the initial state for 2 owners")
	mintValue := big.NewInt(3010)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	validatorOwnerA, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	validatorOwnerB, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	log.Info("working with the following addresses",
		"validatorOwnerA", validatorOwnerA.Bech32, "validatorOwnerB", validatorOwnerB.Bech32)

	log.Info("Step 3. Do 2 stake transactions and test that the new keys are on queue / auction list and have the correct topup")

	topupA := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(100))
	stakeValueA := big.NewInt(0).Add(staking.MinimumStakeValue, topupA)
	txStakeA := generateStakeTransaction(t, cs, validatorOwnerA, blsKeys[0], stakeValueA)

	topupB := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(200))
	stakeValueB := big.NewInt(0).Add(staking.MinimumStakeValue, topupB)
	txStakeB := generateStakeTransaction(t, cs, validatorOwnerB, blsKeys[1], stakeValueB)

	stakeTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txStakeA, txStakeB}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 2, len(stakeTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorOwnerA.Bytes, blsKeys[0], topupA, 2)
	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorOwnerB.Bytes, blsKeys[1], topupB, 2)

	log.Info("Step 4. Convert both validators into staking providers and test that the new keys are on queue / auction list and have the correct topup")

	txConvertA := generateConvertToStakingProviderTransaction(t, cs, validatorOwnerA)
	txConvertB := generateConvertToStakingProviderTransaction(t, cs, validatorOwnerB)

	convertTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txConvertA, txConvertB}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 2, len(convertTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	delegationAddressA := convertTxs[0].Logs.Events[0].Topics[1]
	delegationAddressB := convertTxs[1].Logs.Events[0].Topics[1]

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddressA, blsKeys[0], topupA, 2)
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddressB, blsKeys[1], topupB, 2)

	log.Info("Step 5. If the staking v4 is activated, check that the auction list sorted the 2 BLS keys based on topup")
	step1ActivationEpoch := metachainNode.GetCoreComponents().EnableEpochsHandler().GetActivationEpoch(common.StakingV4Step1Flag)
	if step1ActivationEpoch > metachainNode.GetCoreComponents().EnableEpochsHandler().GetCurrentEpoch() {
		// we are in staking v3.5, the test ends here
		return
	}

	auctionList, err := metachainNode.GetProcessComponents().ValidatorsProvider().GetAuctionList()
	require.Nil(t, err)

	firstAuctionPosition := auctionList[0]
	secondAuctionPosition := auctionList[1]
	// check the correct order of the nodes in the auction list based on topup
	require.Equal(t, blsKeys[1], firstAuctionPosition.Nodes[0].BlsKey)
	require.Equal(t, topupB.String(), firstAuctionPosition.TopUpPerNode)

	require.Equal(t, blsKeys[0], secondAuctionPosition.Nodes[0].BlsKey)
	require.Equal(t, topupA.String(), secondAuctionPosition.TopUpPerNode)
}

// Test description:
// Test that 1 contract having 3 BLS keys proper handles the stakeNodes-unstakeNodes-unBondNodes sequence for 2 of the BLS keys
// 1. Add 3 new validator private keys in the multi key handler
// 2. Set the initial state for 1 owner and 1 delegator
// 3. Do a stake transaction and test that the new key is on queue / auction list and has the correct topup
// 4. Convert the validator into a staking providers and test that the key is on queue / auction list and has the correct topup
// 5. Add 2 nodes in the staking contract
// 6. Delegate 5000 EGLD to the contract
// 7. Stake the 2 nodes
// 8. UnStake 2 nodes (latest staked)
// 9. Unbond the 2 nodes (that were un staked)

// Internal test scenario #85
func TestChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    80,
	}

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		// unbond succeeded because the nodes were on queue
		testChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(t, cs, 1, staking.NotStakedStatus)
	})
	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4
				cfg.EpochConfig.EnableEpochs.AlwaysMergeContextsInEEIEnableEpoch = 1

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(t, cs, 2, staking.UnStakedStatus)
	})
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4
				cfg.EpochConfig.EnableEpochs.AlwaysMergeContextsInEEIEnableEpoch = 1

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(t, cs, 3, staking.UnStakedStatus)
	})
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
			AlterConfigsFunction: func(cfg *config.Configs) {
				cfg.EpochConfig.EnableEpochs.StakeLimitsEnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step1EnableEpoch = 2
				cfg.EpochConfig.EnableEpochs.StakingV4Step2EnableEpoch = 3
				cfg.EpochConfig.EnableEpochs.StakingV4Step3EnableEpoch = 4
				cfg.EpochConfig.EnableEpochs.AlwaysMergeContextsInEEIEnableEpoch = 1

				cfg.EpochConfig.EnableEpochs.MaxNodesChangeEnableEpoch[2].EpochEnable = 4
				cfg.SystemSCConfig.StakingSystemSCConfig.NodeLimitPercentage = 1
			},
		})
		require.Nil(t, err)
		require.NotNil(t, cs)

		defer cs.Close()

		testChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(t, cs, 4, staking.UnStakedStatus)
	})
}

func testChainSimulatorMakeNewContractFromValidatorDataWith1StakingContractUnstakeAndUnbond(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	targetEpoch int32,
	nodesStatusAfterUnBondTx string,
) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	log.Info("Step 1. Add 3 new validator private keys in the multi key handler")
	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(3)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	log.Info("Step 2. Set the initial state for 1 owner and 1 delegator")
	mintValue := big.NewInt(10001)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	owner, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	delegator, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	log.Info("working with the following addresses",
		"owner", owner.Bech32, "", delegator.Bech32)

	log.Info("Step 3. Do a stake transaction and test that the new key is on queue / auction list and has the correct topup")

	topup := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(99))
	stakeValue := big.NewInt(0).Add(staking.MinimumStakeValue, topup)
	txStake := generateStakeTransaction(t, cs, owner, blsKeys[0], stakeValue)

	stakeTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txStake}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(stakeTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, owner.Bytes, blsKeys[0], topup, 1)

	log.Info("Step 4. Convert the validator into a staking providers and test that the key is on queue / auction list and has the correct topup")

	txConvert := generateConvertToStakingProviderTransaction(t, cs, owner)

	convertTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txConvert}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(convertTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	assert.Nil(t, err)

	delegationAddress := convertTxs[0].Logs.Events[0].Topics[1]

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], topup, 1)

	log.Info("Step 5. Add 2 nodes in the staking contract")
	txDataFieldAddNodes := fmt.Sprintf("addNodes@%s@%s@%s@%s", blsKeys[1], staking.MockBLSSignature+"02", blsKeys[2], staking.MockBLSSignature+"03")
	ownerNonce := staking.GetNonce(t, cs, owner)
	txAddNodes := staking.GenerateTransaction(owner.Bytes, ownerNonce, delegationAddress, big.NewInt(0), txDataFieldAddNodes, staking.GasLimitForStakeOperation)

	addNodesTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txAddNodes}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(addNodesTxs))

	log.Info("Step 6. Delegate 5000 EGLD to the contract")
	delegateValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(5000))
	txDataFieldDelegate := "delegate"
	delegatorNonce := staking.GetNonce(t, cs, delegator)
	txDelegate := staking.GenerateTransaction(delegator.Bytes, delegatorNonce, delegationAddress, delegateValue, txDataFieldDelegate, staking.GasLimitForStakeOperation)

	delegateTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txDelegate}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(delegateTxs))

	log.Info("Step 7. Stake the 2 nodes")
	txDataFieldStakeNodes := fmt.Sprintf("stakeNodes@%s@%s", blsKeys[1], blsKeys[2])
	ownerNonce = staking.GetNonce(t, cs, owner)
	txStakeNodes := staking.GenerateTransaction(owner.Bytes, ownerNonce, delegationAddress, big.NewInt(0), txDataFieldStakeNodes, staking.GasLimitForStakeOperation)

	stakeNodesTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txStakeNodes}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(stakeNodesTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the nodes
	assert.Nil(t, err)

	// all 3 nodes should be staked (auction list is 1 as there is one delegation SC with 3 BLS keys in the auction list)
	testBLSKeysAreInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys, topup, 1)

	log.Info("Step 8. UnStake 2 nodes (latest staked)")

	txDataFieldUnStakeNodes := fmt.Sprintf("unStakeNodes@%s@%s", blsKeys[1], blsKeys[2])
	ownerNonce = staking.GetNonce(t, cs, owner)
	txUnStakeNodes := staking.GenerateTransaction(owner.Bytes, ownerNonce, delegationAddress, big.NewInt(0), txDataFieldUnStakeNodes, staking.GasLimitForStakeOperation)

	unStakeNodesTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txUnStakeNodes}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(unStakeNodesTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the nodes
	assert.Nil(t, err)

	// all that only one node is staked (auction list is 1 as there is one delegation SC with 1 BLS key in the auction list)
	expectedTopUp := big.NewInt(0)
	expectedTopUp.Add(topup, delegateValue) // 99 + 5000 = 5099
	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], expectedTopUp, 1)

	log.Info("Step 9. Unbond the 2 nodes (that were un staked)")

	txDataFieldUnBondNodes := fmt.Sprintf("unBondNodes@%s@%s", blsKeys[1], blsKeys[2])
	ownerNonce = staking.GetNonce(t, cs, owner)
	txUnBondNodes := staking.GenerateTransaction(owner.Bytes, ownerNonce, delegationAddress, big.NewInt(0), txDataFieldUnBondNodes, staking.GasLimitForStakeOperation)

	unBondNodesTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txUnBondNodes}, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(unBondNodesTxs))

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the nodes
	assert.Nil(t, err)

	keyStatus := staking.GetAllNodeStates(t, metachainNode, delegationAddress)
	require.Equal(t, len(blsKeys), len(keyStatus))
	// key[0] should be staked
	require.Equal(t, staking.StakedStatus, keyStatus[blsKeys[0]])
	// key[1] and key[2] should be unstaked (unbond was not executed)
	require.Equal(t, nodesStatusAfterUnBondTx, keyStatus[blsKeys[1]])
	require.Equal(t, nodesStatusAfterUnBondTx, keyStatus[blsKeys[2]])
}

func generateStakeTransaction(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	owner dtos.WalletAddress,
	blsKeyHex string,
	stakeValue *big.Int,
) *transaction.Transaction {
	account, err := cs.GetAccount(owner)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeyHex, staking.MockBLSSignature)
	return staking.GenerateTransaction(owner.Bytes, account.Nonce, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
}

func generateConvertToStakingProviderTransaction(
	t *testing.T,
	cs chainSimulatorIntegrationTests.ChainSimulator,
	owner dtos.WalletAddress,
) *transaction.Transaction {
	account, err := cs.GetAccount(owner)
	require.Nil(t, err)

	txDataField := fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, hexServiceFee)
	return staking.GenerateTransaction(owner.Bytes, account.Nonce, vm.DelegationManagerSCAddress, staking.ZeroValue, txDataField, gasLimitForConvertOperation)
}

// Test description
//  Test the creation of a new delegation contract, adding nodes to it, delegating, and undelegating.

// Test scenario
// 1. Initialize the chain simulator
// 2. Generate blocks to activate staking phases
// 3. Create a new delegation contract
// 4. Add validator nodes to the delegation contract
// 5. Perform delegation operations
// 6. Perform undelegation operations
// 7. Validate the results at each step
func TestChainSimulator_CreateNewDelegationContract(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test scenario done in staking 3.5 phase (staking v4 is not active)
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is staked
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorCreateNewDelegationContract(t, cs, 1)
	})

	// Test scenario done in staking v4 phase step 1
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorCreateNewDelegationContract(t, cs, 2)
	})

	// Test scenario done in staking v4 phase step 2
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorCreateNewDelegationContract(t, cs, 3)
	})

	// Test scenario done in staking v4 phase step 3
	// 1. Add a new validator private key in the multi key handler
	// 2. Set the initial state for the owner and the 2 delegators
	// 3. Create a new delegation contract with 1250 egld
	// 3. Add node to the delegation contract
	// 4. Execute 2 delegation operations of 1250 EGLD each, check the topup is 3750
	// 5. Stake node, check the topup is 1250, check the node is in action list
	// 5. Execute 2 unDelegate operations of 100 EGLD each, check the topup is back to 1250
	// 6. Check the node is unstaked in the next epoch
	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorCreateNewDelegationContract(t, cs, 4)
	})

}

func testChainSimulatorCreateNewDelegationContract(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	// Create new validator owner and delegators with initial funds
	validatorOwnerBytes := generateWalletAddressBytes()
	validatorOwner, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(validatorOwnerBytes)
	delegator1Bytes := generateWalletAddressBytes()
	delegator1, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegator1Bytes)
	delegator2Bytes := generateWalletAddressBytes()
	delegator2, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegator2Bytes)
	initialFunds := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(10000)) // 10000 EGLD for each
	addresses := []*dtos.AddressState{
		{Address: validatorOwner, Balance: initialFunds.String()},
		{Address: delegator1, Balance: initialFunds.String()},
		{Address: delegator2, Balance: initialFunds.String()},
	}
	err = cs.SetStateMultiple(addresses)
	require.Nil(t, err)

	// Step 3: Create a new delegation contract
	maxDelegationCap := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(51000)) // 51000 EGLD cap
	txCreateDelegationContract := staking.GenerateTransaction(validatorOwnerBytes, 0, vm.DelegationManagerSCAddress, staking.InitialDelegationValue,
		fmt.Sprintf("createNewDelegationContract@%s@%s", hex.EncodeToString(maxDelegationCap.Bytes()), hexServiceFee),
		gasLimitForDelegationContractCreationOperation)
	createDelegationContractTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txCreateDelegationContract, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, createDelegationContractTx)

	// Check delegation contract creation was successful
	data := createDelegationContractTx.SmartContractResults[0].Data
	parts := strings.Split(data, "@")
	require.Equal(t, 3, len(parts))

	require.Equal(t, hex.EncodeToString([]byte("ok")), parts[1])
	delegationContractAddressHex, _ := hex.DecodeString(parts[2])
	delegationContractAddress, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(delegationContractAddressHex)

	output, err := executeQuery(cs, core.MetachainShardId, vm.DelegationManagerSCAddress, "getAllContractAddresses", nil)
	require.Nil(t, err)
	returnAddress, err := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Encode(output.ReturnData[0])
	require.Nil(t, err)
	require.Equal(t, delegationContractAddress, returnAddress)
	delegationContractAddressBytes := output.ReturnData[0]

	// Step 2: Add validator nodes to the delegation contract
	// This step requires generating BLS keys for validators, signing messages, and sending the "addNodes" transaction.
	// Add checks to verify nodes are added successfully.
	validatorSecretKeysBytes, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)
	err = cs.AddValidatorKeys(validatorSecretKeysBytes)
	require.Nil(t, err)

	signatures := getSignatures(delegationContractAddressBytes, validatorSecretKeysBytes)
	txAddNodes := staking.GenerateTransaction(validatorOwnerBytes, 1, delegationContractAddressBytes, staking.ZeroValue, addNodesTxData(blsKeys, signatures), gasLimitForAddNodesOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys := getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 1, len(notStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(notStakedKeys[0]))
	require.Equal(t, 0, len(unStakedKeys))

	expectedTopUp := big.NewInt(0).Set(staking.InitialDelegationValue)
	expectedTotalStaked := big.NewInt(0).Set(staking.InitialDelegationValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{validatorOwnerBytes})
	require.Nil(t, err)
	require.Equal(t, staking.InitialDelegationValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	// Step 3: Perform delegation operations
	txDelegate1 := staking.GenerateTransaction(delegator1Bytes, 0, delegationContractAddressBytes, staking.InitialDelegationValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, staking.InitialDelegationValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, staking.InitialDelegationValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator1Bytes})
	require.Nil(t, err)
	require.Equal(t, staking.InitialDelegationValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	txDelegate2 := staking.GenerateTransaction(delegator2Bytes, 0, delegationContractAddressBytes, staking.InitialDelegationValue, "delegate", gasLimitForDelegate)
	delegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate2, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate2Tx)

	expectedTopUp = expectedTopUp.Add(expectedTopUp, staking.InitialDelegationValue)
	expectedTotalStaked = expectedTotalStaked.Add(expectedTotalStaked, staking.InitialDelegationValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator2Bytes})
	require.Nil(t, err)
	require.Equal(t, staking.InitialDelegationValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	// Step 4: Perform stakeNodes

	txStakeNodes := staking.GenerateTransaction(validatorOwnerBytes, 2, delegationContractAddressBytes, staking.ZeroValue, fmt.Sprintf("stakeNodes@%s", blsKeys[0]), staking.GasLimitForStakeOperation)
	stakeNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStakeNodes, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeNodesTx)

	expectedTopUp = expectedTopUp.Sub(expectedTopUp, staking.InitialDelegationValue)
	expectedTopUp = expectedTopUp.Sub(expectedTopUp, staking.InitialDelegationValue)
	require.Equal(t, expectedTopUp, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Make block finalized
	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationContractAddressBytes, blsKeys[0], expectedTopUp, 1)

	// Step 5: Perform unDelegate from 1 user
	// The nodes should remain in the staked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate1 := staking.GenerateTransaction(delegator1Bytes, 1, delegationContractAddressBytes, staking.ZeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(staking.InitialDelegationValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate1Tx)

	expectedTopUp = expectedTopUp.Sub(expectedTopUp, staking.InitialDelegationValue)
	expectedTotalStaked = expectedTotalStaked.Sub(expectedTotalStaked, staking.InitialDelegationValue)
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, expectedTotalStaked, big.NewInt(0).SetBytes(output.ReturnData[0]))
	require.Equal(t, expectedTopUp.String(), getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes).String())

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator1Bytes})
	require.Nil(t, err)
	require.Equal(t, staking.ZeroValue, big.NewInt(0).SetBytes(output.ReturnData[0]))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	// Step 6: Perform unDelegate from last user
	// The nodes should remain in the unStaked state
	// The total active stake should be reduced by the amount undelegated

	txUndelegate2 := staking.GenerateTransaction(delegator2Bytes, 1, delegationContractAddressBytes, staking.ZeroValue, fmt.Sprintf("unDelegate@%s", hex.EncodeToString(staking.InitialDelegationValue.Bytes())), gasLimitForUndelegateOperation)
	undelegate2Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txUndelegate2, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, undelegate2Tx)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getTotalActiveStake", nil)
	require.Nil(t, err)
	require.Equal(t, "1250000000000000000000", big.NewInt(0).SetBytes(output.ReturnData[0]).String())
	require.Equal(t, staking.ZeroValue, getBLSTopUpValue(t, metachainNode, delegationContractAddressBytes))

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getUserActiveStake", [][]byte{delegator2Bytes})
	require.Nil(t, err)
	require.Equal(t, "0", big.NewInt(0).SetBytes(output.ReturnData[0]).String())

	// still staked until epoch change
	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 1, len(stakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(stakedKeys[0]))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 0, len(unStakedKeys))

	err = cs.GenerateBlocksUntilEpochIsReached(targetEpoch + 1)
	require.Nil(t, err)

	output, err = executeQuery(cs, core.MetachainShardId, delegationContractAddressBytes, "getAllNodeStates", nil)
	require.Nil(t, err)
	stakedKeys, notStakedKeys, unStakedKeys = getNodesFromContract(output.ReturnData)
	require.Equal(t, 0, len(stakedKeys))
	require.Equal(t, 0, len(notStakedKeys))
	require.Equal(t, 1, len(unStakedKeys))
	require.Equal(t, blsKeys[0], hex.EncodeToString(unStakedKeys[0]))
}

func generateWalletAddressBytes() []byte {
	buff := make([]byte, walletAddressBytesLen)
	_, _ = rand.Read(buff)

	return buff
}

func executeQuery(cs chainSimulatorIntegrationTests.ChainSimulator, shardID uint32, scAddress []byte, funcName string, args [][]byte) (*dataVm.VMOutputApi, error) {
	output, _, err := cs.GetNodeHandler(shardID).GetFacadeHandler().ExecuteSCQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  funcName,
		Arguments: args,
	})
	return output, err
}

func addNodesTxData(blsKeys []string, sigs [][]byte) string {
	txData := "addNodes"

	for i := range blsKeys {
		txData = txData + "@" + blsKeys[i] + "@" + hex.EncodeToString(sigs[i])
	}

	return txData
}

func getSignatures(msg []byte, blsKeys [][]byte) [][]byte {
	signer := mclsig.NewBlsSigner()

	signatures := make([][]byte, len(blsKeys))
	for i, blsKey := range blsKeys {
		sk, _ := signing.NewKeyGenerator(mcl.NewSuiteBLS12()).PrivateKeyFromByteArray(blsKey)
		signatures[i], _ = signer.Sign(sk, msg)
	}

	return signatures
}

func getNodesFromContract(returnData [][]byte) ([][]byte, [][]byte, [][]byte) {
	var stakedKeys, notStakedKeys, unStakedKeys [][]byte

	for i := 0; i < len(returnData); i += 2 {
		switch string(returnData[i]) {
		case "staked":
			stakedKeys = append(stakedKeys, returnData[i+1])
		case "notStaked":
			notStakedKeys = append(notStakedKeys, returnData[i+1])
		case "unStaked":
			unStakedKeys = append(unStakedKeys, returnData[i+1])
		}
	}
	return stakedKeys, notStakedKeys, unStakedKeys
}

func getBLSTopUpValue(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, address []byte) *big.Int {
	scQuery := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStakedTopUpStakedBlsKeys",
		CallerAddr: vm.StakingSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{address},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, staking.OkReturnCode, result.ReturnCode)

	if len(result.ReturnData[0]) == 0 {
		return big.NewInt(0)
	}

	return big.NewInt(0).SetBytes(result.ReturnData[0])
}

// Test description:
// Test that merging delegation  with whiteListForMerge and mergeValidatorToDelegationWithWhitelist contracts still works properly
// Test that their topups will merge too and will be used by auction list computing.
//
// Internal test scenario #12
func TestChainSimulator_MergeDelegation(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    30,
	}

	// Test steps:
	// 1. User A - Stake 1 node to have 100 egld more than minimum required stake value
	// 2. User A - Execute `makeNewContractFromValidatorData` to create delegation contract based on User A account
	// 3. User B - Stake 1 node with more than 2500 egld
	// 4. User A - Execute `whiteListForMerge@addressA` in order to whitelist for merge User B
	// 5. User B - Execute `mergeValidatorToDelegationWithWhitelist@delegationContract` in order to merge User B to delegation contract created at step 2.

	t.Run("staking ph 4 is not active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMergingDelegation(t, cs, 1)
	})

	t.Run("staking ph 4 step 1 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMergingDelegation(t, cs, 2)
	})

	t.Run("staking ph 4 step 2 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMergingDelegation(t, cs, 3)
	})

	t.Run("staking ph 4 step 3 is active", func(t *testing.T) {
		cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
			BypassTxSignatureCheck:   false,
			TempDir:                  t.TempDir(),
			PathToInitialConfig:      defaultPathToInitialConfig,
			NumOfShards:              3,
			GenesisTimestamp:         time.Now().Unix(),
			RoundDurationInMillis:    roundDurationInMillis,
			RoundsPerEpoch:           roundsPerEpoch,
			ApiInterface:             api.NewNoApiInterface(),
			MinNodesPerShard:         3,
			MetaChainMinNodes:        3,
			NumNodesWaitingListMeta:  3,
			NumNodesWaitingListShard: 3,
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

		testChainSimulatorMergingDelegation(t, cs, 4)
	})
}

func testChainSimulatorMergingDelegation(t *testing.T, cs chainSimulatorIntegrationTests.ChainSimulator, targetEpoch int32) {
	err := cs.GenerateBlocksUntilEpochIsReached(targetEpoch)
	require.Nil(t, err)

	privateKeys, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(3)
	require.Nil(t, err)

	err = cs.AddValidatorKeys(privateKeys)
	require.Nil(t, err)
	metachainNode := cs.GetNodeHandler(core.MetachainShardId)

	mintValue := big.NewInt(3000)
	mintValue = mintValue.Mul(staking.OneEGLD, mintValue)

	validatorA, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	validatorB, err := cs.GenerateAndMintWalletAddress(core.AllShardId, mintValue)
	require.Nil(t, err)

	log.Info("Step 1. User A: - stake 1 node to have 100 egld more than minimum stake value")
	stakeValue := big.NewInt(0).Set(staking.MinimumStakeValue)
	addedStakedValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(100))
	stakeValue.Add(stakeValue, addedStakedValue)
	txDataField := fmt.Sprintf("stake@01@%s@%s", blsKeys[0], staking.MockBLSSignature)
	txStake := staking.GenerateTransaction(validatorA.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorA.Bytes, blsKeys[0], addedStakedValue, 1)
	require.Equal(t, addedStakedValue, getBLSTopUpValue(t, metachainNode, validatorA.Bytes))

	log.Info("Step 2. Execute MakeNewContractFromValidatorData for User A")
	txDataField = fmt.Sprintf("makeNewContractFromValidatorData@%s@%s", maxCap, hexServiceFee)
	txConvert := staking.GenerateTransaction(validatorA.Bytes, 1, vm.DelegationManagerSCAddress, staking.ZeroValue, txDataField, gasLimitForConvertOperation)
	convertTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	delegationAddress := convertTx.Logs.Events[0].Topics[1]

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, delegationAddress, blsKeys[0], addedStakedValue, 1)

	log.Info("Step 3. User B: - stake 1 node to have 100 egld more")
	stakeValue = big.NewInt(0).Set(staking.MinimumStakeValue)
	addedStakedValue = big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(100))
	stakeValue.Add(stakeValue, addedStakedValue)
	txDataField = fmt.Sprintf("stake@01@%s@%s", blsKeys[1], staking.MockBLSSignature)
	txStake = staking.GenerateTransaction(validatorB.Bytes, 0, vm.ValidatorSCAddress, stakeValue, txDataField, staking.GasLimitForStakeOperation)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	testBLSKeyIsInQueueOrAuction(t, metachainNode, validatorB.Bytes, blsKeys[1], addedStakedValue, 2)
	require.Equal(t, addedStakedValue, getBLSTopUpValue(t, metachainNode, validatorB.Bytes))

	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	require.Equal(t, delegationAddress, getBLSKeyOwner(t, metachainNode, decodedBLSKey0))

	decodedBLSKey1, _ := hex.DecodeString(blsKeys[1])
	require.Equal(t, validatorB.Bytes, getBLSKeyOwner(t, metachainNode, decodedBLSKey1))

	log.Info("Step 4. User A : whitelistForMerge@addressB")
	txDataField = fmt.Sprintf("whitelistForMerge@%s", hex.EncodeToString(validatorB.Bytes))
	whitelistForMerge := staking.GenerateTransaction(validatorA.Bytes, 2, delegationAddress, staking.ZeroValue, txDataField, gasLimitForDelegate)
	whitelistForMergeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(whitelistForMerge, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, whitelistForMergeTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	log.Info("Step 5. User A : mergeValidatorToDelegationWithWhitelist")
	txDataField = fmt.Sprintf("mergeValidatorToDelegationWithWhitelist@%s", hex.EncodeToString(delegationAddress))

	txConvert = staking.GenerateTransaction(validatorB.Bytes, 1, vm.DelegationManagerSCAddress, staking.ZeroValue, txDataField, gasLimitForMergeOperation)
	convertTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txConvert, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, convertTx)

	err = cs.GenerateBlocks(2) // allow the metachain to finalize the block that contains the staking of the node
	require.Nil(t, err)

	decodedBLSKey0, _ = hex.DecodeString(blsKeys[0])
	require.Equal(t, delegationAddress, getBLSKeyOwner(t, metachainNode, decodedBLSKey0))

	decodedBLSKey1, _ = hex.DecodeString(blsKeys[1])
	require.Equal(t, delegationAddress, getBLSKeyOwner(t, metachainNode, decodedBLSKey1))

	expectedTopUpValue := big.NewInt(0).Mul(staking.OneEGLD, big.NewInt(200))
	require.Equal(t, expectedTopUpValue, getBLSTopUpValue(t, metachainNode, delegationAddress))
}

func getBLSKeyOwner(t *testing.T, metachainNode chainSimulatorProcess.NodeHandler, blsKey []byte) []byte {
	scQuery := &process.SCQuery{
		ScAddress:  vm.StakingSCAddress,
		FuncName:   "getOwner",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{blsKey},
	}
	result, _, err := metachainNode.GetFacadeHandler().ExecuteSCQuery(scQuery)
	require.Nil(t, err)
	require.Equal(t, staking.OkReturnCode, result.ReturnCode)

	return result.ReturnData[0]
}
