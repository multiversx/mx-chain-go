//go:build !race
// +build !race

package delegation

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegationSystemClaimMulti(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(10000) // 10%
	numContracts := 2
	totalNumNodes := 5
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	ownerAddresses := getAddresses(numContracts)

	delegators := getAddresses(numDelegators)
	delegationScAddresses := make([][]byte, numContracts)

	firstTwoDelegators := delegators[:2]
	lastTwoDelegators := delegators[2:]
	halfDelegationVal := delegationVal / 2

	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(2000), ownerAddresses[i])
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{ownerAddresses[i]}, delegationScAddresses[i], big.NewInt(2000))

		blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddresses[i], totalNumNodes)
		txData := addNodesTxData(blsKeys, sigs)
		returnedCode, err := processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)

		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], "delegate", big.NewInt(delegationVal))
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[i], big.NewInt(delegationVal))

		processMultipleTransactions(t, tpn, lastTwoDelegators, delegationScAddresses[i], "delegate", big.NewInt(halfDelegationVal))
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[i], big.NewInt(halfDelegationVal))

		txData = txDataForFunc("stakeNodes", blsKeys)
		returnedCode, err = processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}

	verifyValidatorSCStake(t, tpn, delegationScAddresses[0], big.NewInt(5000))
	verifyValidatorSCStake(t, tpn, delegationScAddresses[1], big.NewInt(5000))

	for i := range delegationScAddresses {
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(1000), 1, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(2000), 2, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(3000), 3, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(4000), 4, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(5000), 5, 1)
	}

	for i := range delegationScAddresses {
		checkRewardData(t, tpn, delegationScAddresses[i], 1, 1000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 2, 2000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 3, 3000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 4, 4000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 5, 5000, 5000, serviceFee)
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[0], 2700)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[1], 2700)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6900)
	}

	txData := "claimMulti"
	listAddresses := make([][]byte, 0)
	for _, address := range delegationScAddresses {
		txData += "@" + hex.EncodeToString(address)
		listAddresses = append(listAddresses, address)
	}

	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[0], "claimMulti", listAddresses, 5400)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[1], "claimMulti", listAddresses, 5400)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[2], "claimMulti", listAddresses, 2700)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[3], "claimMulti", listAddresses, 2700)

	for _, delegator := range delegators {
		tpn.ScrForwarder.CreateBlockStarted()
		returnedCode, err := processTransaction(tpn, delegator, vm.DelegationManagerSCAddress, txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
		assert.Equal(t, len(delegationScAddresses), len(tpn.ScrForwarder.GetAllCurrentFinishedTxs()))
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[0], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[1], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6900)
	}

	verifyValidatorSCStake(t, tpn, delegationScAddresses[0], big.NewInt(5000))
	verifyValidatorSCStake(t, tpn, delegationScAddresses[1], big.NewInt(5000))
}

func TestDelegationSystemRedelegateMulti(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000000)
	serviceFee := big.NewInt(10000) // 10%
	numContracts := 2
	totalNumNodes := 5
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	ownerAddresses := getAddresses(numContracts)

	delegators := getAddresses(numDelegators)
	delegationScAddresses := make([][]byte, numContracts)

	firstTwoDelegators := delegators[:2]
	lastTwoDelegators := delegators[2:]
	halfDelegationVal := delegationVal / 2

	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(2000), ownerAddresses[i])
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{ownerAddresses[i]}, delegationScAddresses[i], big.NewInt(2000))

		blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddresses[i], totalNumNodes)
		txData := addNodesTxData(blsKeys, sigs)
		returnedCode, err := processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)

		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], "delegate", big.NewInt(delegationVal))
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[i], big.NewInt(delegationVal))

		processMultipleTransactions(t, tpn, lastTwoDelegators, delegationScAddresses[i], "delegate", big.NewInt(halfDelegationVal))
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[i], big.NewInt(halfDelegationVal))

		txData = txDataForFunc("stakeNodes", blsKeys)
		returnedCode, err = processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}

	verifyValidatorSCStake(t, tpn, delegationScAddresses[0], big.NewInt(5000))
	verifyValidatorSCStake(t, tpn, delegationScAddresses[1], big.NewInt(5000))

	for i := range delegationScAddresses {
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(1000), 1, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(2000), 2, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(3000), 3, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(4000), 4, 1)
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(5000), 5, 1)
	}

	for i := range delegationScAddresses {
		checkRewardData(t, tpn, delegationScAddresses[i], 1, 1000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 2, 2000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 3, 3000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 4, 4000, 5000, serviceFee)
		checkRewardData(t, tpn, delegationScAddresses[i], 5, 5000, 5000, serviceFee)
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[0], 2700)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[1], 2700)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6900)
	}

	txData := "reDelegateMulti"
	listAddresses := make([][]byte, 0)
	for _, address := range delegationScAddresses {
		txData += "@" + hex.EncodeToString(address)
		listAddresses = append(listAddresses, address)
	}

	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[0], "reDelegateMulti", listAddresses, 5400)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[1], "reDelegateMulti", listAddresses, 5400)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[2], "reDelegateMulti", listAddresses, 2700)
	checkClaimMultiReturn(t, tpn, vm.DelegationManagerSCAddress, delegators[3], "reDelegateMulti", listAddresses, 2700)

	for _, delegator := range delegators {
		returnedCode, err := processTransaction(tpn, delegator, vm.DelegationManagerSCAddress, txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[0], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[1], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 0)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6900)
	}

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[0], big.NewInt(3700))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[1], big.NewInt(3700))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[0], big.NewInt(1850))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[1], big.NewInt(1850))

	verifyValidatorSCStake(t, tpn, delegationScAddresses[0], big.NewInt(5000+8100))
	verifyValidatorSCStake(t, tpn, delegationScAddresses[1], big.NewInt(5000+8100))
}

func checkClaimMultiReturn(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	delegMgrAddr []byte,
	delegAddr []byte,
	function string,
	arguments [][]byte,
	expectedRewards int64,
) {
	query := &process.SCQuery{
		ScAddress:  delegMgrAddr,
		FuncName:   function,
		CallerAddr: delegAddr,
		CallValue:  big.NewInt(0),
		Arguments:  arguments,
	}
	vmOutput, _, err := tpn.SCQueryService.ExecuteQuery(query)
	assert.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	require.Equal(t, big.NewInt(expectedRewards).Bytes(), vmOutput.ReturnData[0])
}
