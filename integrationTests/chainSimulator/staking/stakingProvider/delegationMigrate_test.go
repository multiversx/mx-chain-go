package stakingProvider

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
	chainSimulatorIntegrationTests "github.com/multiversx/mx-chain-go/integrationTests/chainSimulator"
	"github.com/multiversx/mx-chain-go/integrationTests/chainSimulator/staking"
	"github.com/multiversx/mx-chain-go/node/chainSimulator"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/components/api"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/vm"

	"github.com/stretchr/testify/require"
)

// create 2 staking providers (sp1 and sp2)
// delegate 500 EGLD from a new address
// migrate 100 EGLD from sp1 to sp2 should work
// migrate 100 EGLD from sp2 to sp1 should NOT WORK (is in cool down period)
// delegate 100 EGLD from first address to sp2
func TestMigrateStakingProviderHappyFlow(t *testing.T) {
	roundDurationInMillis := uint64(6000)
	roundsPerEpoch := core.OptionalUint64{
		HasValue: true,
		Value:    20,
	}

	cs, err := chainSimulator.NewChainSimulator(chainSimulator.ArgsChainSimulator{
		BypassTxSignatureCheck: true,
		TempDir:                t.TempDir(),
		PathToInitialConfig:    defaultPathToInitialConfig,
		NumOfShards:            3,
		GenesisTimestamp:       time.Now().Unix(),
		RoundDurationInMillis:  roundDurationInMillis,
		RoundsPerEpoch:         roundsPerEpoch,
		ApiInterface:           api.NewNoApiInterface(),
		MinNodesPerShard:       1,
		MetaChainMinNodes:      1,
		AlterConfigsFunction: func(cfg *config.Configs) {
			configs.SetStakingV4ActivationEpochs(cfg, 2)
		},
	})
	require.Nil(t, err)
	require.NotNil(t, cs)
	defer cs.Close()

	mintValue := big.NewInt(0).Mul(big.NewInt(5000), chainSimulatorIntegrationTests.OneEGLD)
	sp1Owner, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	sp2Owner, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	delegator, err := cs.GenerateAndMintWalletAddress(0, mintValue)
	require.Nil(t, err)

	err = cs.GenerateBlocks(1)
	require.Nil(t, err)

	err = cs.GenerateBlocksUntilEpochIsReached(1)
	require.Nil(t, err)

	// create sp1
	stakeValue, _ := big.NewInt(0).SetString("4250000000000000000000", 10)
	dataField := "createNewDelegationContract@00@0ea1"
	txStake := chainSimulatorIntegrationTests.GenerateTransaction(sp1Owner.Bytes, staking.GetNonce(t, cs, sp1Owner), vm.DelegationManagerSCAddress, stakeValue, dataField, 80_000_000)
	stakeTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	delegationAddress := stakeTx.Logs.Events[2].Address
	sp1AddressBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegationAddress)

	txStake = chainSimulatorIntegrationTests.GenerateTransaction(sp2Owner.Bytes, staking.GetNonce(t, cs, sp2Owner), vm.DelegationManagerSCAddress, stakeValue, dataField, 80_000_000)
	stakeTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txStake, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, stakeTx)

	delegationAddress = stakeTx.Logs.Events[2].Address
	sp2AddressBytes, _ := cs.GetNodeHandler(0).GetCoreComponents().AddressPubKeyConverter().Decode(delegationAddress)

	fmt.Println(sp1AddressBytes, sp2AddressBytes)

	// do delegate to sp1
	delegateValue := big.NewInt(0).Mul(big.NewInt(500), chainSimulatorIntegrationTests.OneEGLD)
	txDelegate1 := chainSimulatorIntegrationTests.GenerateTransaction(delegator.Bytes, 0, sp1AddressBytes, delegateValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	err = cs.ChangeEpochs(40)
	require.Nil(t, err)

	// move 100 egld from sp1 --> sp2 should work
	moveValue := big.NewInt(0).Mul(big.NewInt(100), chainSimulatorIntegrationTests.OneEGLD)
	dataField = fmt.Sprintf("changeStakingProvider@%s@%s@%s", hex.EncodeToString(moveValue.Bytes()), hex.EncodeToString(sp1AddressBytes), hex.EncodeToString(sp2AddressBytes))
	tx := chainSimulatorIntegrationTests.GenerateTransaction(delegator.Bytes, staking.GetNonce(t, cs, delegator), vm.DelegationManagerSCAddress, big.NewInt(0), dataField, 80_000_000)
	moveDelegationTx, err := cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, moveDelegationTx)

	// move second time 100 egld from sp1 --> sp2 should work
	moveValue = big.NewInt(0).Mul(big.NewInt(100), chainSimulatorIntegrationTests.OneEGLD)
	dataField = fmt.Sprintf("changeStakingProvider@%s@%s@%s", hex.EncodeToString(moveValue.Bytes()), hex.EncodeToString(sp1AddressBytes), hex.EncodeToString(sp2AddressBytes))
	tx = chainSimulatorIntegrationTests.GenerateTransaction(delegator.Bytes, staking.GetNonce(t, cs, delegator), vm.DelegationManagerSCAddress, big.NewInt(0), dataField, 80_000_000)
	moveDelegationTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, moveDelegationTx)

	// delegate 100 EGLD to sp2
	delegateValue = big.NewInt(0).Mul(big.NewInt(100), chainSimulatorIntegrationTests.OneEGLD)
	txDelegate1 = chainSimulatorIntegrationTests.GenerateTransaction(delegator.Bytes, staking.GetNonce(t, cs, delegator), sp2AddressBytes, delegateValue, "delegate", gasLimitForDelegate)
	delegate1Tx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(txDelegate1, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, delegate1Tx)

	// move 100 egld from sp2 --> sp1 should NOT work (is in cool down period)
	moveValue = big.NewInt(0).Mul(big.NewInt(100), chainSimulatorIntegrationTests.OneEGLD)
	dataField = fmt.Sprintf("changeStakingProvider@%s@%s@%s", hex.EncodeToString(moveValue.Bytes()), hex.EncodeToString(sp2AddressBytes), hex.EncodeToString(sp1AddressBytes))
	tx = chainSimulatorIntegrationTests.GenerateTransaction(delegator.Bytes, staking.GetNonce(t, cs, delegator), vm.DelegationManagerSCAddress, big.NewInt(0), dataField, 80_000_000)
	moveDelegationTx, err = cs.SendTxAndGenerateBlockTilTxIsExecuted(tx, staking.MaxNumOfBlockToGenerateWhenExecutingTx)
	require.Nil(t, err)
	require.NotNil(t, moveDelegationTx)

	eventWithError := moveDelegationTx.Logs.Events[0]
	require.Equal(t, eventWithError.Identifier, core.SignalErrorOperation)
	require.Equal(t, "cannot migrate to another service provider during cooldown period, wait 30 more epochs", string(eventWithError.Topics[1]))

	// check delegation values for delegator in sp1 and sp2
	output, err := executeQuery(cs, core.MetachainShardId, sp1AddressBytes, "getUserActiveStake", [][]byte{delegator.Bytes})
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0).Mul(big.NewInt(300), chainSimulatorIntegrationTests.OneEGLD), big.NewInt(0).SetBytes(output.ReturnData[0]))

	output, err = executeQuery(cs, core.MetachainShardId, sp2AddressBytes, "getUserActiveStake", [][]byte{delegator.Bytes})
	require.Nil(t, err)
	require.Equal(t, big.NewInt(0).Mul(big.NewInt(300), chainSimulatorIntegrationTests.OneEGLD), big.NewInt(0).SetBytes(output.ReturnData[0]))
}
