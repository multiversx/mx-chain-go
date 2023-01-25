//go:build !race
// +build !race

package delegation

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/mcl"
	mclsig "github.com/multiversx/mx-chain-crypto-go/signing/mcl/singlesig"
	"github.com/multiversx/mx-chain-go/dataRetriever/dataPool"
	"github.com/multiversx/mx-chain-go/integrationTests"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/state"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/vm"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegationSystemNodesOperationsTest(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	value := big.NewInt(1000)

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, value, tpn.OwnAccount.Address)

	scrsHandler, _ := tpn.InterimProcContainer.Get(block.SmartContractResultBlock)
	scrs := scrsHandler.GetAllCurrentFinishedTxs()

	numExpectedScrsFound := 0
	// we expect the following 2 scrs:
	// Delegation Manager -> Delegation Contract Address - 1000 EGLD
	// Delegation Contract Address -> Staking Address - 1000 EGLD
	for _, scr := range scrs {
		if bytes.Equal(scr.GetSndAddr(), vm.DelegationManagerSCAddress) && bytes.Equal(scr.GetRcvAddr(), delegationScAddress) && scr.GetValue().Cmp(value) == 0 {
			numExpectedScrsFound++
		}

		if bytes.Equal(scr.GetSndAddr(), delegationScAddress) && bytes.Equal(scr.GetRcvAddr(), vm.ValidatorSCAddress) && scr.GetValue().Cmp(value) == 0 {
			numExpectedScrsFound++
		}
	}

	assert.Equal(t, 2, numExpectedScrsFound)
}

func TestDelegationSystemNodesOperations(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 7
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 7 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// remove 2 nodes from the delegation contract
	numNodesToStake := totalNumNodes - 2
	txData = txDataForFunc("removeNodes", blsKeys[numNodesToStake:])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))

	// stake 5 nodes
	txData = txDataForFunc("stakeNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake], "staked")

	// unStake 3 nodes
	txData = txDataForFunc("unStakeNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake], "unStaked")

	// remove nodes should fail because they are not unBonded
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake])
	returnedCode, _ = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.UserError, returnedCode)

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 10000000})
	// unBond nodes
	txData = txDataForFunc("unBondNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	txData = "unDelegate@" + hex.EncodeToString(big.NewInt(1000).Bytes())
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// remove unBonded nodes should work
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)
}

func TestDelegationSystemReStakeNodes(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 7
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 7 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// remove 2 nodes from the delegation contract
	numNodesToStake := totalNumNodes - 2
	txData = txDataForFunc("removeNodes", blsKeys[numNodesToStake:])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))

	// stake 5 nodes
	txData = txDataForFunc("stakeNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake], "staked")

	// unStake 3 nodes
	txData = txDataForFunc("unStakeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake-3], "unStaked")

	processMultipleTransactions(t, tpn, delegators[:1], delegationScAddress, "unDelegate@"+hex.EncodeToString(big.NewInt(delegationVal).Bytes()), big.NewInt(0))
	// remove nodes should fail because they are not unBonded
	txData = txDataForFunc("reStakeUnStakedNodes", blsKeys[:numNodesToStake-4])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake-4], "staked")
}

func TestDelegationChangeConfig(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	numDelegators := 4
	initialDelegationValue := int64(100)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(initialDelegationValue))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(initialDelegationValue))

	// increase the min delegation stake value
	newMinDelegationAmount := big.NewInt(101)
	changeAddress, err := integrationTests.TestAddressPubkeyConverter.Decode(integrationTests.DelegationManagerConfigChangeAddress)
	require.Nil(t, err)
	txData := fmt.Sprintf("changeMinDelegationAmount@%s", hex.EncodeToString(newMinDelegationAmount.Bytes()))
	retCode, err := processTransaction(tpn, changeAddress, vm.DelegationManagerSCAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, retCode)

	// check that the min delegation amount has the correct value
	scQuery := &process.SCQuery{
		ScAddress:  vm.DelegationManagerSCAddress,
		FuncName:   "getContractConfig",
		CallerAddr: vm.DelegationManagerSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  make([][]byte, 0),
	}
	vmOutput, err := tpn.SCQueryService.ExecuteQuery(scQuery)
	require.Nil(t, err)
	assert.Equal(t, newMinDelegationAmount.Bytes(), vmOutput.ReturnData[5])

	// try delegate with initialDelegationValue, should fail
	newDelegator := getAddresses(1)[0]
	retCode, err = processTransaction(tpn, newDelegator, delegationScAddress, "delegate", big.NewInt(initialDelegationValue))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.UserError, retCode)

	// try delegate with the new value
	delegators = getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", newMinDelegationAmount)
}

func TestDelegationSystemDelegateUnDelegateFromTopUpWithdraw(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 3 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))

	// stake 3 nodes
	txData = txDataForFunc("stakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// unDelegate all from 2 delegators
	txData = "unDelegate" + "@" + intToString(uint32(delegationVal))
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, txData, big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal))
	verifyUserUndelegatedList(t, tpn, delegationScAddress, delegators[0], []*big.Int{big.NewInt(delegationVal)})
	// withdraw unDelegated delegators should not withdraw because of unBond period
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 1})

	// withdraw unDelegated delegators should withdraw after unBond period has passed
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorIsDeleted(t, tpn, delegators[:numDelegators-2], delegationScAddress)
}

func TestDelegationSystemDelegateUnDelegateOnlyPartOfDelegation(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 3 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))

	// stake 3 nodes
	txData = txDataForFunc("stakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// unDelegate half from delegators
	txData = "unDelegate" + "@" + intToString(uint32(delegationVal/2))
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, txData, big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))

	// withdraw unDelegated delegators should not withdraw because of unBond period
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 1})

	// withdraw unDelegated delegators should withdraw after unBond period has passed
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
}

func TestDelegationSystemMultipleDelegationContractsAndSameBlsKeysShouldNotWork(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	numContracts := 2
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	ownerAddresses := getAddresses(numContracts)
	for i := range ownerAddresses {
		integrationTests.MintAddress(tpn.AccntState, ownerAddresses[i], big.NewInt(2000))
	}

	// create 2 new delegation contracts
	delegationScAddresses := make([][]byte, numContracts)
	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), ownerAddresses[i])
	}

	// add same BLS keys to all delegation contracts
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	signer := mclsig.NewBlsSigner()

	pubKeys := make([][]byte, totalNumNodes)
	secretKeys := make([]crypto.PrivateKey, totalNumNodes)
	for i := 0; i < totalNumNodes; i++ {
		sk, pk := keyGen.GeneratePair()
		pubKeys[i], _ = pk.ToByteArray()
		secretKeys[i] = sk
	}

	for i := range delegationScAddresses {
		sigs := make([][]byte, totalNumNodes)
		for j := range secretKeys {
			sigs[j], _ = signer.Sign(secretKeys[j], delegationScAddresses[i])
		}

		txData := addNodesTxData(pubKeys, sigs)
		returnedCode, err := processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}

	// 4 delegators fill the delegation cap for each contract
	delegators := getAddresses(numDelegators)

	for i := range delegationScAddresses {
		processMultipleTransactions(t, tpn, delegators, delegationScAddresses[i], "delegate", big.NewInt(delegationVal))
	}

	for i := range delegationScAddresses {
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddresses[i], big.NewInt(delegationVal))
	}

	// stake 3 nodes for each contract
	txData := txDataForFunc("stakeNodes", pubKeys)
	returnedCode, err := processTransaction(tpn, ownerAddresses[0], delegationScAddresses[0], txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	returnedCode, _ = processTransaction(tpn, ownerAddresses[1], delegationScAddresses[1], txData, big.NewInt(0))
	assert.Equal(t, vmcommon.UserError, returnedCode)
}

func TestDelegationSystemMultipleDelegationContractsAndSameDelegators(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	numContracts := 2
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	ownerAddresses := getAddresses(numContracts)
	for i := range ownerAddresses {
		integrationTests.MintAddress(tpn.AccntState, ownerAddresses[i], big.NewInt(2000))
	}

	delegators := getAddresses(numDelegators)
	delegationScAddresses := make([][]byte, numContracts)

	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), ownerAddresses[i])

		blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddresses[i], totalNumNodes)
		txData := addNodesTxData(blsKeys, sigs)
		returnedCode, err := processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)

		processMultipleTransactions(t, tpn, delegators, delegationScAddresses[i], "delegate", big.NewInt(delegationVal))

		verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddresses[i], big.NewInt(delegationVal))

		txData = txDataForFunc("stakeNodes", blsKeys)
		returnedCode, err = processTransaction(tpn, ownerAddresses[i], delegationScAddresses[i], txData, big.NewInt(0))
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}

	firstTwoDelegators := delegators[:2]

	for i := range delegationScAddresses {
		txData := "unDelegate" + "@" + intToString(uint32(delegationVal))
		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], txData, big.NewInt(0))
	}

	for i := range delegationScAddresses {
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[i], big.NewInt(0))
		verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", firstTwoDelegators, delegationScAddresses[i], big.NewInt(delegationVal))
	}

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 1})

	for i := range delegationScAddresses {
		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], "withdraw", big.NewInt(0))
	}

	for i := range delegationScAddresses {
		verifyDelegatorIsDeleted(t, tpn, firstTwoDelegators, delegationScAddresses[i])
	}
}

func TestDelegationRewardsComputationAfterChangeServiceFee(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(500000)
	serviceFee := big.NewInt(10000) // 10%
	totalNumNodes := 5
	numDelegators := 4
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 5 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	firstDelegators := delegators[:2]
	firstDelegatorsValue := big.NewInt(1500)
	lastDelegators := delegators[2:]
	lastDelegatorsValue := big.NewInt(500)
	processMultipleTransactions(t, tpn, firstDelegators, delegationScAddress, "delegate", firstDelegatorsValue)
	processMultipleTransactions(t, tpn, lastDelegators, delegationScAddress, "delegate", lastDelegatorsValue)

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstDelegators, delegationScAddress, firstDelegatorsValue)
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastDelegators, delegationScAddress, lastDelegatorsValue)

	// stake 5 nodes
	txData = txDataForFunc("stakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(1000), 1, 1)
	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(2000), 2, 1)

	txData = "changeServiceFee@" + hex.EncodeToString(big.NewInt(20000).Bytes()) // 20%
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.Ok, returnedCode)
	assert.Nil(t, err)

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(1000), 3, 1)
	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(2000), 4, 1)

	checkRewardData(t, tpn, delegationScAddress, 1, 1000, 5000, serviceFee)
	checkRewardData(t, tpn, delegationScAddress, 2, 2000, 5000, serviceFee)
	checkRewardData(t, tpn, delegationScAddress, 3, 1000, 5000, big.NewInt(20000))
	checkRewardData(t, tpn, delegationScAddress, 4, 2000, 5000, big.NewInt(20000))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Nonce: 5, Epoch: 4})

	checkDelegatorReward(t, tpn, delegationScAddress, delegators[0], 1530)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[1], 1530)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[2], 510)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[3], 510)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 1920)

	txData = "unDelegate@" + hex.EncodeToString(firstDelegatorsValue.Bytes())
	returnedCode, err = processTransaction(tpn, delegators[0], delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.Ok, returnedCode)
	assert.Nil(t, err)
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:1], delegationScAddress, big.NewInt(0))

	// claim rewards for the same user multiple times - should return only one
	txData = "claimRewards"
	for _, delegator := range delegators {
		returnedCode, err = processTransaction(tpn, delegator, delegationScAddress, txData, big.NewInt(0))
		assert.Equal(t, vmcommon.Ok, returnedCode)
		assert.Nil(t, err)
	}

	checkDelegatorReward(t, tpn, delegationScAddress, delegators[0], 0)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[1], 0)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[2], 0)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[3], 0)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 1920)

	for _, delegator := range delegators {
		returnedCode, err = processTransaction(tpn, delegator, delegationScAddress, txData, big.NewInt(0))
		assert.Equal(t, vmcommon.Ok, returnedCode)
		assert.Nil(t, err)
	}

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(1000))

	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, "reDelegateRewards", big.NewInt(0))
	assert.Equal(t, vmcommon.Ok, returnedCode)
	assert.Nil(t, err)

	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 0)
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(2920))
}

func TestDelegationUnJail(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(10000) // 10%
	totalNumNodes := 5
	numDelegators := 4
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 5 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	firstDelegators := delegators[:2]
	firstDelegatorsValue := big.NewInt(1500)
	lastDelegators := delegators[2:]
	lastDelegatorsValue := big.NewInt(500)
	processMultipleTransactions(t, tpn, firstDelegators, delegationScAddress, "delegate", firstDelegatorsValue)
	processMultipleTransactions(t, tpn, lastDelegators, delegationScAddress, "delegate", lastDelegatorsValue)

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstDelegators, delegationScAddress, firstDelegatorsValue)
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastDelegators, delegationScAddress, lastDelegatorsValue)

	// stake 5 nodes
	txData = txDataForFunc("stakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// jail 2 bls keys
	txData = txDataForFunc("jail", blsKeys[:2])
	returnedCode, err = processTransaction(tpn, vm.JailingAddress, vm.StakingSCAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:2], "jailed")
	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[2:], "staked")

	// unJail blsKeys
	txData = txDataForFunc("unJailNodes", blsKeys[:2])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(20))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:2], "staked")
	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[2:], "staked")
}

func TestDelegationSystemDelegateSameUsersAFewTimes(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(0)
	serviceFee := big.NewInt(00)
	totalNumNodes := 1
	numDelegators := 2
	delegationVal := int64(5000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	validatorAcc := getAsUserAccount(tpn, vm.ValidatorSCAddress)
	genesisBalance := validatorAcc.GetBalance()

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1350), tpn.OwnAccount.Address)

	// add 1 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// set automatic activation on
	txData = "setAutomaticActivation@" + hex.EncodeToString([]byte("true"))
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// self delegate 1250 eGLD
	txData = "delegate"
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(1250))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// delegators delegate
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(2600))

	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal*2))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(2600))

	verifyValidatorSCStake(t, tpn, delegationScAddress, big.NewInt(22600))
	delegationAcc := getAsUserAccount(tpn, delegationScAddress)
	assert.Equal(t, delegationAcc.GetBalance(), big.NewInt(0))

	validatorAcc = getAsUserAccount(tpn, vm.ValidatorSCAddress)
	assert.Equal(t, validatorAcc.GetBalance(), big.NewInt(0).Add(genesisBalance, big.NewInt(22600)))
}

func TestDelegationSystemMultipleDelegationContractsAndSameDelegatorsClaimRewardsMultipleTimeUndelegateClaimRewardsMultipleTime(t *testing.T) {
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
	quarterDelegationVal := halfDelegationVal / 2

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

	txData := "claimRewards"
	for i := 0; i < 10; i++ {
		for j := 0; j < 2; j++ {
			returnedCode, err := processTransaction(tpn, delegators[j], delegationScAddresses[0], txData, big.NewInt(0))
			assert.Equal(t, vmcommon.Ok, returnedCode)
			assert.Nil(t, err)
		}
	}

	checkDelegatorReward(t, tpn, delegationScAddresses[0], delegators[0], 0)
	checkDelegatorReward(t, tpn, delegationScAddresses[0], delegators[1], 0)
	checkDelegatorReward(t, tpn, delegationScAddresses[1], delegators[0], 2700)
	checkDelegatorReward(t, tpn, delegationScAddresses[1], delegators[1], 2700)

	for i := range delegationScAddresses {
		verifyValidatorSCStake(t, tpn, delegationScAddresses[i], big.NewInt(5000))
	}

	for i := range delegationScAddresses {
		txData = "unDelegate" + "@" + intToString(uint32(delegationVal))
		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], txData, big.NewInt(0))
	}

	for i := range delegationScAddresses {
		verifyValidatorSCStake(t, tpn, delegationScAddresses[i], big.NewInt(3000))
	}

	for i := range delegationScAddresses {
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[i], big.NewInt(0))
		verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", firstTwoDelegators, delegationScAddresses[i], big.NewInt(delegationVal))
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[i], big.NewInt(halfDelegationVal))
		verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", lastTwoDelegators, delegationScAddresses[i], big.NewInt(0))
	}

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 5, Nonce: 50})

	for i := range delegationScAddresses {
		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], "withdraw", big.NewInt(0))

		txData = "unDelegate" + "@" + intToString(uint32(quarterDelegationVal))
		processMultipleTransactions(t, tpn, lastTwoDelegators, delegationScAddresses[i], txData, big.NewInt(0))
	}

	for i := range delegationScAddresses {
		verifyValidatorSCStake(t, tpn, delegationScAddresses[i], big.NewInt(2500))
	}

	for i := range delegationScAddresses {
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", lastTwoDelegators, delegationScAddresses[i], big.NewInt(quarterDelegationVal))
		verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", lastTwoDelegators, delegationScAddresses[i], big.NewInt(quarterDelegationVal))
	}

	txData = "claimRewards"
	for i := 0; i < 10; i++ {
		for j := 0; j < 2; j++ {
			returnedCode, err := processTransaction(tpn, delegators[j], delegationScAddresses[1], txData, big.NewInt(0))
			assert.Equal(t, vmcommon.Ok, returnedCode)
			assert.Nil(t, err)
		}
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 1350)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6900)
	}

	for i := range delegationScAddresses {
		addRewardsToDelegation(tpn, delegationScAddresses[i], big.NewInt(100), 6, 50)
	}

	for i := range delegationScAddresses {
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[2], 1359)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], delegators[3], 1359)
		checkDelegatorReward(t, tpn, delegationScAddresses[i], ownerAddresses[i], 6982)
	}
}

func TestDelegationSystemDelegateUnDelegateReceiveRewardsWhenAllIsUndelegated(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(10000) // 10%
	totalNumNodes := 2
	numDelegators := 2
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(3000), tpn.OwnAccount.Address)

	// add 2 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 2 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(3000))

	// stake 2 nodes
	txData = txDataForFunc("stakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(200), 1, 1)
	checkRewardData(t, tpn, delegationScAddress, 1, 200, 5000, serviceFee)

	checkDelegatorReward(t, tpn, delegationScAddress, delegators[0], 36)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[1], 36)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 128)

	// unDelegate all from delegators
	txData = "unDelegate" + "@" + intToString(uint32(delegationVal))
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, txData, big.NewInt(0))

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(100), 2, 1)
	checkRewardData(t, tpn, delegationScAddress, 2, 100, 3000, serviceFee)

	checkDelegatorReward(t, tpn, delegationScAddress, delegators[0], 36)
	checkDelegatorReward(t, tpn, delegationScAddress, delegators[1], 36)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 228)

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators, delegationScAddress, big.NewInt(delegationVal))
	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(3000))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(0))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 2, Nonce: 50})

	// withdraw unDelegated delegators
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators, delegationScAddress, big.NewInt(0))

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(300), 3, 50)
	checkRewardData(t, tpn, delegationScAddress, 3, 300, 3000, serviceFee)

	txData = "claimRewards"
	for i := 0; i < 2; i++ {
		returnedCode, err = processTransaction(tpn, delegators[i], delegationScAddress, txData, big.NewInt(0))
		assert.Equal(t, vmcommon.Ok, returnedCode)
		assert.Nil(t, err)
	}

	verifyDelegatorIsDeleted(t, tpn, delegators, delegationScAddress)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 528)

	// unStake 2 nodes
	txData = txDataForFunc("unStakeNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 3, Nonce: 100})

	// unBond 2 nodes
	txData = txDataForFunc("unBondNodes", blsKeys)
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(400), 4, 100)
	checkRewardData(t, tpn, delegationScAddress, 4, 400, 3000, serviceFee)

	verifyDelegatorIsDeleted(t, tpn, delegators, delegationScAddress)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 928)

	// unDelegate all from owner
	txData = "unDelegate" + "@" + intToString(uint32(3000))
	processMultipleTransactions(t, tpn, [][]byte{tpn.OwnAccount.Address}, delegationScAddress, txData, big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(3000))

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(500), 5, 100)
	checkRewardData(t, tpn, delegationScAddress, 5, 500, 0, serviceFee)

	verifyDelegatorIsDeleted(t, tpn, delegators, delegationScAddress)
	checkDelegatorReward(t, tpn, delegationScAddress, tpn.OwnAccount.Address, 928)

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: 5, Nonce: 150})

	// withdraw unDelegated owner
	processMultipleTransactions(t, tpn, [][]byte{tpn.OwnAccount.Address}, delegationScAddress, "withdraw", big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", [][]byte{tpn.OwnAccount.Address}, delegationScAddress, big.NewInt(0))

	txData = "claimRewards"
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.Ok, returnedCode)
	assert.Nil(t, err)

	verifyDelegatorIsDeleted(t, tpn, delegators, delegationScAddress)

	addRewardsToDelegation(tpn, delegationScAddress, big.NewInt(600), 6, 150)
	checkRewardData(t, tpn, delegationScAddress, 6, 600, 0, serviceFee)

	verifyDelegatorIsDeleted(t, tpn, delegators, delegationScAddress)
}

func TestDelegationSystemCleanUpContract(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(integrationTests.ArgTestProcessorNode{
		MaxShards:            1,
		NodeShardId:          core.MetachainShardId,
		TxSignPrivKeyShardId: 0,
	})
	tpn.InitDelegationManager()
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 7
	numDelegators := 4
	tpn.EpochNotifier.CheckEpoch(&testscommon.HeaderHandlerStub{
		EpochField: integrationTests.UnreachableEpoch + 1,
	})
	delegationVal := int64(1000)

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 1})

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1000), tpn.OwnAccount.Address)

	// add 7 nodes to the delegation contract
	blsKeys, sigs := getBlsKeysAndSignatures(delegationScAddress, totalNumNodes)
	txData := addNodesTxData(blsKeys, sigs)
	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// remove 2 nodes from the delegation contract
	numNodesToStake := totalNumNodes - 2
	txData = txDataForFunc("removeNodes", blsKeys[numNodesToStake:])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// 4 delegators fill the delegation cap
	delegators := getAddresses(numDelegators)
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, "delegate", big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators, delegationScAddress, big.NewInt(delegationVal))

	// stake 5 nodes
	txData = txDataForFunc("stakeNodes", blsKeys[:numNodesToStake])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake], "staked")

	// unStake 3 nodes
	txData = txDataForFunc("unStakeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	checkNodesStatus(t, tpn, vm.StakingSCAddress, blsKeys[:numNodesToStake-3], "unStaked")

	// remove nodes should fail because they are not unBonded
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, _ = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.UserError, returnedCode)

	tpn.BlockchainHook.SetCurrentHeader(&block.MetaBlock{Nonce: 10000000})
	// unBond nodes
	txData = txDataForFunc("unBondNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	// remove unBonded nodes should work
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)
}

func verifyUserUndelegatedList(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	scAddress []byte,
	delegator []byte,
	values []*big.Int,
) {
	query := &process.SCQuery{
		ScAddress:  scAddress,
		FuncName:   "getUserUnDelegatedList",
		CallerAddr: delegator,
		Arguments:  [][]byte{delegator},
		CallValue:  big.NewInt(0),
	}
	vmOutput, err := tpn.SCQueryService.ExecuteQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	assert.Equal(t, len(values)*2, len(vmOutput.ReturnData))

	for i, returnData := range vmOutput.ReturnData {
		if i*2 >= len(values) {
			break
		}
		value := values[i*2]
		assert.Equal(t, value.Bytes(), returnData)
	}
}

func getAsUserAccount(node *integrationTests.TestProcessorNode, address []byte) state.UserAccountHandler {
	acc, _ := node.AccntState.GetExistingAccount(address)
	userAcc, _ := acc.(state.UserAccountHandler)
	return userAcc
}

func verifyValidatorSCStake(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	delegationAddr []byte,
	expectedRes *big.Int,
) {
	query := &process.SCQuery{
		ScAddress:  vm.ValidatorSCAddress,
		FuncName:   "getTotalStaked",
		CallerAddr: vm.ValidatorSCAddress,
		CallValue:  big.NewInt(0),
		Arguments:  [][]byte{delegationAddr},
	}
	vmOutput, err := tpn.SCQueryService.ExecuteQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	assert.Equal(t, string(vmOutput.ReturnData[0]), expectedRes.String())
}

func checkNodesStatus(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	destAddr []byte,
	blsKeys [][]byte,
	expectedStatus string,
) {
	for i := range blsKeys {
		nodeStatus := viewFuncSingleResult(t, tpn, destAddr, "getBLSKeyStatus", [][]byte{blsKeys[i]})
		assert.Equal(t, expectedStatus, string(nodeStatus))
	}
}

func checkDelegatorReward(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	delegScAddr []byte,
	delegAddr []byte,
	expectedRewards int64,
) {
	delegRewards := viewFuncSingleResult(t, tpn, delegScAddr, "getClaimableRewards", [][]byte{delegAddr})
	assert.Equal(t, big.NewInt(expectedRewards).Bytes(), delegRewards)
}

func checkRewardData(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	delegScAddr []byte,
	epoch uint8,
	expectedRewards int64,
	expectedTotalActive int64,
	expectedServiceFee *big.Int,
) {
	epoch0RewardData := viewFuncMultipleResults(t, tpn, delegScAddr, "getRewardData", [][]byte{{epoch}})
	assert.Equal(t, big.NewInt(expectedRewards).Bytes(), epoch0RewardData[0])
	assert.Equal(t, big.NewInt(expectedTotalActive).Bytes(), epoch0RewardData[1])
	assert.Equal(t, expectedServiceFee.Bytes(), epoch0RewardData[2])
}

func addRewardsToDelegation(tpn *integrationTests.TestProcessorNode, recvAddr []byte, value *big.Int, epoch uint32, nonce uint64) {
	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Epoch: epoch, Nonce: nonce})

	tx := &rewardTx.RewardTx{
		Round:   0,
		Value:   value,
		RcvAddr: recvAddr,
		Epoch:   0,
	}
	rewardTxSerialized, _ := integrationTests.TestMarshalizer.Marshal(tx)
	rewardTxHash := integrationTests.TestHasher.Compute(string(rewardTxSerialized))

	mbSlice := block.MiniBlockSlice{
		&block.MiniBlock{
			TxHashes:        [][]byte{rewardTxHash},
			ReceiverShardID: core.MetachainShardId,
			Type:            block.RewardsBlock,
		},
	}

	txCacher := dataPool.NewCurrentBlockTransactionsPool()
	txCacher.AddTx(rewardTxHash, tx)

	_ = tpn.EpochStartSystemSCProcessor.ProcessDelegationRewards(mbSlice, txCacher)
}

func verifyDelegatorsStake(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	funcName string,
	addresses [][]byte,
	delegationAddr []byte,
	expectedRes *big.Int,
) {
	for i := range addresses {
		delegActiveStake := viewFuncSingleResult(t, tpn, delegationAddr, funcName, [][]byte{addresses[i]})
		assert.Equal(t, expectedRes, big.NewInt(0).SetBytes(delegActiveStake))
	}
}

func verifyDelegatorIsDeleted(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	addresses [][]byte,
	delegationAddr []byte,
) {
	for _, address := range addresses {
		query := &process.SCQuery{
			ScAddress:  delegationAddr,
			FuncName:   "isDelegator",
			CallerAddr: vm.EndOfEpochAddress,
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{address},
		}
		vmOutput, err := tpn.SCQueryService.ExecuteQuery(query)
		assert.Nil(t, err)
		assert.Equal(t, vmOutput.ReturnMessage, "view function works only for existing delegators")
		assert.Equal(t, vmOutput.ReturnCode, vmcommon.UserError)
	}
}

func deployNewSc(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	maxDelegationCap *big.Int,
	serviceFee *big.Int,
	value *big.Int,
	ownerAddress []byte,
) []byte {
	txData := "createNewDelegationContract" + "@" + hex.EncodeToString(maxDelegationCap.Bytes()) + "@" + hex.EncodeToString(serviceFee.Bytes())
	returnedCode, err := processTransaction(tpn, ownerAddress, vm.DelegationManagerSCAddress, txData, value)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	logs := tpn.TransactionLogProcessor.GetAllCurrentLogs()
	tpn.TransactionLogProcessor.Clean()

	for _, log := range logs {
		for _, event := range log.GetLogEvents() {
			if string(event.GetIdentifier()) == "writeLog" && bytes.Equal(event.GetAddress(), vm.DelegationManagerSCAddress) {
				tokens := strings.Split(string(event.GetData()), "@")
				address, _ := hex.DecodeString(tokens[2])
				return address
			}
		}
	}

	return []byte{}
}

func viewFuncSingleResult(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	address []byte,
	function string,
	arguments [][]byte,
) []byte {
	returnData := getReturnDataFromQuery(t, tpn, address, function, arguments)
	return returnData[0]
}

func viewFuncMultipleResults(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	address []byte,
	function string,
	arguments [][]byte,
) [][]byte {
	return getReturnDataFromQuery(t, tpn, address, function, arguments)
}

func getReturnDataFromQuery(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	address []byte,
	function string,
	arguments [][]byte,
) [][]byte {
	query := &process.SCQuery{
		ScAddress:  address,
		FuncName:   function,
		CallerAddr: vm.EndOfEpochAddress,
		CallValue:  big.NewInt(0),
		Arguments:  arguments,
	}
	vmOutput, err := tpn.SCQueryService.ExecuteQuery(query)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)

	return vmOutput.ReturnData
}

func intToString(val uint32) string {
	valueToUnDelegate := make([]byte, 4)
	binary.BigEndian.PutUint32(valueToUnDelegate, val)
	return hex.EncodeToString(valueToUnDelegate)
}

func processMultipleTransactions(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	delegatorsAddr [][]byte,
	receiverAddr []byte,
	txData string,
	value *big.Int,
) {
	for i := range delegatorsAddr {
		returnedCode, err := processTransaction(tpn, delegatorsAddr[i], receiverAddr, txData, value)
		assert.Nil(t, err)
		assert.Equal(t, vmcommon.Ok, returnedCode)
	}
}

func txDataForFunc(function string, blsKeys [][]byte) string {
	txData := function

	for i := range blsKeys {
		txData = txData + "@" + hex.EncodeToString(blsKeys[i])
	}

	return txData
}

func addNodesTxData(blsKeys, sigs [][]byte) string {
	txData := "addNodes"

	for i := range blsKeys {
		txData = txData + "@" + hex.EncodeToString(blsKeys[i]) + "@" + hex.EncodeToString(sigs[i])
	}

	return txData
}

func getBlsKeysAndSignatures(msg []byte, num int) ([][]byte, [][]byte) {
	keyGen := signing.NewKeyGenerator(mcl.NewSuiteBLS12())
	signer := mclsig.NewBlsSigner()

	pubKeys := make([][]byte, num)
	signatures := make([][]byte, num)
	for i := 0; i < num; i++ {
		sk, pk := keyGen.GeneratePair()
		signatures[i], _ = signer.Sign(sk, msg)
		pubKeys[i], _ = pk.ToByteArray()
	}

	return pubKeys, signatures
}

func getAddresses(num int) [][]byte {
	addresses := make([][]byte, num)
	for i := 0; i < num; i++ {
		addresses[i] = integrationTests.CreateRandomBytes(32)
	}

	return addresses
}

func processTransaction(
	tpn *integrationTests.TestProcessorNode,
	senderAddr []byte,
	receiverAddr []byte,
	txData string,
	value *big.Int,
) (vmcommon.ReturnCode, error) {
	tx := &transaction.Transaction{
		Nonce:    tpn.OwnAccount.Nonce,
		Value:    value,
		SndAddr:  senderAddr,
		RcvAddr:  receiverAddr,
		Data:     []byte(txData),
		GasPrice: integrationTests.MinTxGasPrice,
		GasLimit: integrationTests.MinTxGasLimit + uint64(len(txData)) + integrationTests.AdditionalGasLimit,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
	}

	return tpn.TxProcessor.ProcessTransaction(tx)
}
