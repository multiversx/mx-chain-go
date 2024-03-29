package staking

import (
	"encoding/hex"
	"fmt"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/stretchr/testify/require"
	"math/big"
	"testing"
	"time"
)

func TestStakingProviderWithNodes(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	stakingV4ActivationEpoch := uint32(2)

	t.Run("staking ph 4 step 1 active", func(t *testing.T) {
		testStakingProviderWithNodesReStakeUnStaked(t, stakingV4ActivationEpoch)
	})

	t.Run("staking ph 4 step 2 active", func(t *testing.T) {
		testStakingProviderWithNodesReStakeUnStaked(t, stakingV4ActivationEpoch+1)
	})

	t.Run("staking ph 4 step 3 active", func(t *testing.T) {
		testStakingProviderWithNodesReStakeUnStaked(t, stakingV4ActivationEpoch+2)
	})
}

func testStakingProviderWithNodesReStakeUnStaked(t *testing.T, stakingV4ActivationEpoch uint32) {
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

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
			configs.SetStakingV4ActivationEpochs(cfg, stakingV4ActivationEpoch)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	mintValue := big.NewInt(0).Mul(big.NewInt(5000), oneEGLD)
	validatorOwner, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	// create delegation contract
	stakeValue, _ := big.NewInt(0).SetString("4250000000000000000000", 10)
	dataField := "createNewDelegationContract@00@0ea1"
	txStake := generateTransaction(validatorOwner.Bytes, getNonce(t, cs, validatorOwner), vm.DelegationManagerSCAddress, stakeValue, dataField, 80_000_000)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	delegationAddress := stakeTx.Logs.Events[2].Address
	delegationAddressBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegationAddress)

	// add nodes in queue
	_, blsKeys, err := chainSimulator.GenerateBlsPrivateKeys(1)
	require.Nil(t, err)

	txDataFieldAddNodes := fmt.Sprintf("addNodes@%s@%s", blsKeys[0], mockBLSSignature+"02")
	ownerNonce := getNonce(t, cs, validatorOwner)
	txAddNodes := generateTransaction(validatorOwner.Bytes, ownerNonce, delegationAddressBytes, big.NewInt(0), txDataFieldAddNodes, gasLimitForStakeOperation)
	addNodesTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txAddNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, addNodesTx)

	txDataFieldStakeNodes := fmt.Sprintf("stakeNodes@%s", blsKeys[0])
	ownerNonce = getNonce(t, cs, validatorOwner)
	txStakeNodes := generateTransaction(validatorOwner.Bytes, ownerNonce, delegationAddressBytes, big.NewInt(0), txDataFieldStakeNodes, gasLimitForStakeOperation)

	stakeNodesTxs, err := cs.SendTxsAndGenerateBlocksTilAreExecuted([]*transaction.Transaction{txStakeNodes}, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.Equal(t, 1, len(stakeNodesTxs))

	metachainNode := cs.GetNodeHandler(core.MetachainShardId)
	decodedBLSKey0, _ := hex.DecodeString(blsKeys[0])
	status := getBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, "queued", status)

	// activate staking v4
	err = cs.GenerateBlocksUntilEpochIsReached(int32(stakingV4ActivationEpoch))
	require.Nil(t, err)

	status = getBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, "unStaked", status)

	ownerNonce = getNonce(t, cs, validatorOwner)
	reStakeTxData := fmt.Sprintf("reStakeUnStakedNodes@%s", blsKeys[0])
	reStakeNodes := generateTransaction(validatorOwner.Bytes, ownerNonce, delegationAddressBytes, big.NewInt(0), reStakeTxData, gasLimitForStakeOperation)
	reStakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(reStakeNodes, maxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, reStakeTx)

	status = getBLSKeyStatus(t, metachainNode, decodedBLSKey0)
	require.Equal(t, "staked", status)

	err = cs.GenerateBlocks(20)

	checkValidatorStatus(t, cs, blsKeys[0], "auction")
}
