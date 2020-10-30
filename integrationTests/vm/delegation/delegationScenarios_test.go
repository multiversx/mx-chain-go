package delegation

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/crypto/signing"
	"github.com/ElrondNetwork/elrond-go/crypto/signing/mcl"
	mclsig "github.com/ElrondNetwork/elrond-go/crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	"github.com/ElrondNetwork/elrond-go/vm"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestDelegationSystemNodesOperations(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, "node addr")
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 7
	numDelegators := 4
	delegationVal := int64(1000)

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1100), tpn.OwnAccount.Address)

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

	//unStake 3 nodes
	txData = txDataForFunc("unStakeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	//remove nodes should fail because they are not unBonded
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, _ = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Equal(t, vmcommon.UserError, returnedCode)

	//unBond nodes
	txData = txDataForFunc("unBondNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	//remove unBonded nodes should work
	txData = txDataForFunc("removeNodes", blsKeys[:numNodesToStake-3])
	returnedCode, err = processTransaction(tpn, tpn.OwnAccount.Address, delegationScAddress, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)
}

func TestDelegationSystemDelegateUnDelegateFromTopUpWithdraw(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, "node addr")
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(100000001)

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1100), tpn.OwnAccount.Address)

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

	//unDelegate all from 2 delegators
	txData = "unDelegate" + "@" + intToString(uint32(delegationVal))
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, txData, big.NewInt(delegationVal))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal))

	//withdraw unDelegated delegators should not withdraw because of unBond period
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(1000))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Nonce: 5})

	//withdraw unDelegated delegators should withdraw after unBond period has passed
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(1000))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
}

func TestDelegationSystemDelegateUnDelegateOnlyPartOfDelegation(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, "node addr")
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(100000001)

	// create new delegation contract
	delegationScAddress := deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1100), tpn.OwnAccount.Address)

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

	//unDelegate half from delegators
	txData = "unDelegate" + "@" + intToString(uint32(delegationVal/2))
	processMultipleTransactions(t, tpn, delegators, delegationScAddress, txData, big.NewInt(0))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))

	//withdraw unDelegated delegators should not withdraw because of unBond period
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(1000))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Nonce: 5})

	//withdraw unDelegated delegators should withdraw after unBond period has passed
	processMultipleTransactions(t, tpn, delegators[:numDelegators-2], delegationScAddress, "withdraw", big.NewInt(1000))

	verifyDelegatorsStake(t, tpn, "getUserActiveStake", delegators[:numDelegators-2], delegationScAddress, big.NewInt(delegationVal/2))
	verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", delegators[:numDelegators-2], delegationScAddress, big.NewInt(0))
}

func TestDelegationSystemMultipleDelegationContractsAndSameBlsKeysShouldNotWork(t *testing.T) {
	tpn := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, "node addr")
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	numContracts := 2
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)

	ownerAddresses := getAddresses(numContracts)
	for i := range ownerAddresses {
		integrationTests.MintAddress(tpn.AccntState, ownerAddresses[i], big.NewInt(2000))
	}

	// create 2 new delegation contracts
	delegationScAddresses := make([][]byte, numContracts)
	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1100), ownerAddresses[i])
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
	tpn := integrationTests.NewTestProcessorNode(1, core.MetachainShardId, 0, "node addr")
	maxDelegationCap := big.NewInt(5000)
	serviceFee := big.NewInt(1000)
	numContracts := 2
	totalNumNodes := 3
	numDelegators := 4
	delegationVal := int64(1000)
	tpn.EpochNotifier.CheckEpoch(100000001)

	ownerAddresses := getAddresses(numContracts)
	for i := range ownerAddresses {
		integrationTests.MintAddress(tpn.AccntState, ownerAddresses[i], big.NewInt(2000))
	}

	delegators := getAddresses(numDelegators)
	delegationScAddresses := make([][]byte, numContracts)

	for i := range delegationScAddresses {
		delegationScAddresses[i] = deployNewSc(t, tpn, maxDelegationCap, serviceFee, big.NewInt(1100), ownerAddresses[i])

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

	tpn.BlockchainHook.SetCurrentHeader(&block.Header{Nonce: 5})

	for i := range delegationScAddresses {
		processMultipleTransactions(t, tpn, firstTwoDelegators, delegationScAddresses[i], "withdraw", big.NewInt(delegationVal))
	}

	for i := range delegationScAddresses {
		verifyDelegatorsStake(t, tpn, "getUserActiveStake", firstTwoDelegators, delegationScAddresses[i], big.NewInt(0))
		verifyDelegatorsStake(t, tpn, "getUserUnStakedValue", firstTwoDelegators, delegationScAddresses[i], big.NewInt(0))
	}
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
		txData := txDataForFunc(funcName, [][]byte{addresses[i]})
		delegActiveStake := delegatorViewFuncSingleResult(t, tpn, txData, delegationAddr)
		assert.Equal(t, expectedRes, big.NewInt(0).SetBytes(delegActiveStake))
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
	scrForwarder, _ := tpn.ScrForwarder.(interface {
		CleanIntermediateTransactions()
	})
	scrForwarder.CleanIntermediateTransactions()

	txData := "createNewDelegationContract" + "@" + hex.EncodeToString(maxDelegationCap.Bytes()) + "@" + hex.EncodeToString(serviceFee.Bytes())
	returnedCode, err := processTransaction(tpn, ownerAddress, vm.DelegationManagerSCAddress, txData, value)
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	scrs := smartContract.GetAllSCRs(tpn.ScProcessor)
	for i := range scrs {
		tx, isScr := scrs[i].(*smartContractResult.SmartContractResult)
		if !isScr {
			continue
		}

		if bytes.Equal(tx.RcvAddr, ownerAddress) {
			tokens := strings.Split(string(tx.GetData()), "@")
			address, _ := hex.DecodeString(tokens[2])
			return address
		}
	}

	return []byte{}
}

func delegatorViewFuncSingleResult(
	t *testing.T,
	tpn *integrationTests.TestProcessorNode,
	txData string,
	delegScAddr []byte,
) []byte {
	scrForwarder, _ := tpn.ScrForwarder.(interface {
		CleanIntermediateTransactions()
	})
	scrForwarder.CleanIntermediateTransactions()

	returnedCode, err := processTransaction(tpn, tpn.OwnAccount.Address, delegScAddr, txData, big.NewInt(0))
	assert.Nil(t, err)
	assert.Equal(t, vmcommon.Ok, returnedCode)

	scr := smartContract.GetAllSCRs(tpn.ScProcessor)
	if len(scr) != 1 {
		return []byte{}
	}

	tokens := strings.Split(string(scr[0].GetData()), "@")
	res, _ := hex.DecodeString(tokens[2])

	return res
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
