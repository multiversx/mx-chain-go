package upgrades

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestUpgrades_Hello(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = true
	err := context.DeploySC("../testdata/hello-v1/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2")

	err = context.UpgradeSC("../testdata/hello-v2/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v3")

	err = context.UpgradeSC("../testdata/hello-v3/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, "forty-two", context.QuerySCString("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_HelloDoesNotUpgradeWhenNotUpgradeable(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = false
	err := context.DeploySC("../testdata/hello-v1/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2 will not be performed")

	err = context.UpgradeSC("../testdata/hello-v2/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_HelloUpgradesToNotUpgradeable(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = true
	err := context.DeploySC("../testdata/hello-v1/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2, becomes not upgradeable")

	context.ScCodeMetadata.Upgradeable = false
	err = context.UpgradeSC("../testdata/hello-v2/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v3, should not be possible")

	err = context.UpgradeSC("../testdata/hello-v3/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_ParentAndChildContracts(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	var parentAddress []byte
	var childAddress []byte
	owner := &context.Owner

	fmt.Println("Deploy parent")

	err := context.DeploySC("../testdata/upgrades-parent/output/parent.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(45), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
	parentAddress = context.ScAddress

	fmt.Println("Deploy child v1")

	childInitialCode := arwen.GetSCCode("../testdata/hello-v1/output/answer.wasm")
	err = context.ExecuteSC(owner, "createChild@"+childInitialCode)
	require.Nil(t, err)

	fmt.Println("Aquire child address, do query")

	childAddress = context.QuerySCBytes("getChildAddress", [][]byte{})
	context.ScAddress = childAddress
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Deploy child v2")
	context.ScAddress = parentAddress
	// We need to double hex-encode the code (so that we don't have to hex-encode in the contract).
	childUpgradedCode := arwen.GetSCCode("../testdata/hello-v2/output/answer.wasm")
	childUpgradedCode = hex.EncodeToString([]byte(childUpgradedCode))
	// Not supported at this moment, V2 not deployed.
	// TODO: Adjust test when upgrade child from parent is supported.
	err = context.ExecuteSC(owner, "upgradeChild@"+childUpgradedCode)
	require.Nil(t, err)
}

func TestUpgrades_UpgradeDelegationContract(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@0064@0064@0064"
	delegationUpgradeParams := "0000000000000000000000000000000000000000000000000000000000000000@0080@0080@0080"

	context.GasLimit = 21700000
	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(t, err)

	err = context.UpgradeSC(delegationWasmPath, delegationUpgradeParams)
	require.Nil(t, err)
}

func TestUpgrades_DelegationCannotBeUpgradedByNonOwner(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@0064@0064@0064"
	delegationUpgradeParams := "0000000000000000000000000000000000000000000000000000000000000000@0080@0080@0080"

	context.GasLimit = 21700000
	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(t, err)

	context.Owner = context.Alice
	err = context.UpgradeSC(delegationWasmPath, delegationUpgradeParams)
	require.Equal(t, fmt.Errorf("upgrade not allowed"), err)
}

func TestUpgrades_HelloCannotBeUpgradedByNonOwner(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = true
	err := context.DeploySC("../testdata/hello-v1/output/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2 will not be performed")

	// Alice states that she is the owner of the contract (though she is not)
	context.Owner = context.Alice
	err = context.UpgradeSC("../testdata/hello-v2/output/answer.wasm", "")
	require.Equal(t, fmt.Errorf("upgrade not allowed"), err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_CounterCannotBeUpgradedByNonOwner(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	context.ScCodeMetadata.Upgradeable = true
	err := context.DeploySC("../testdata/counter/output/counter.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(1), context.QuerySCInt("get", [][]byte{}))

	err = context.ExecuteSC(&context.Alice, "increment")
	require.Nil(t, err)
	require.Equal(t, uint64(2), context.QuerySCInt("get", [][]byte{}))

	// Alice states that she is the owner of the contract (though she is not)
	// Neither code, nor storage get modified
	context.Owner = context.Alice
	err = context.UpgradeSC("../testdata/counter/output/counter.wasm", "")
	require.Equal(t, fmt.Errorf("upgrade not allowed"), err)
	require.Equal(t, uint64(2), context.QuerySCInt("get", [][]byte{}))
}

func TestUpgrades_TrialAndError(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	numOfShards := 1
	nodesPerShard := 1
	numMetachainNodes := 0

	advertiser := integrationTests.CreateMessengerWithKadDht("")
	_ = advertiser.Bootstrap()

	nodes := integrationTests.CreateNodes(
		numOfShards,
		nodesPerShard,
		numMetachainNodes,
		integrationTests.GetConnectableAddress(advertiser),
	)

	idxProposers := []int{0, 1}
	integrationTests.DisplayAndStartNodes(nodes)

	defer func() {
		_ = advertiser.Close()
		for _, n := range nodes {
			_ = n.Messenger.Close()
		}
	}()

	initialVal := big.NewInt(10000000000000)
	integrationTests.MintAllNodes(nodes, initialVal)

	round := uint64(0)
	nonce := uint64(0)
	round = integrationTests.IncrementAndPrintRound(round)
	nonce++

	alice := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bob := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	integrationTests.MintAddress(nodes[0].AccntState, alice, initialVal)
	integrationTests.MintAddress(nodes[0].AccntState, bob, initialVal)

	deployTxData := fmt.Sprintf("%s@%s@0100", arwen.GetSCCode("../testdata/hello-v1/output/answer.wasm"), hex.EncodeToString(factory.ArwenVirtualMachine))
	upgradeTxData := fmt.Sprintf("upgradeContract@%s@0100", arwen.GetSCCode("../testdata/hello-v2/output/answer.wasm"))

	// Deploy the smart contract. Alice is the owner
	addTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  make([]byte, 32),
		SndAddr:  alice,
		GasPrice: nodes[0].EconomicsData.GetMinGasPrice(),
		GasLimit: nodes[0].EconomicsData.MaxGasLimitPerBlock(0) - 1,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
		Data:     []byte(deployTxData),
	}, nodes)

	scAddress, _ := nodes[0].BlockchainHook.NewAddress(alice, 0, factory.ArwenVirtualMachine)
	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	require.Equal(t, []byte{24}, query(t, nodes[0], scAddress, "getUltimateAnswer"))

	// Upgrade as Bob - upgrade should fail, since Alice is the owner
	addTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  bob,
		GasPrice: nodes[0].EconomicsData.GetMinGasPrice(),
		GasLimit: nodes[0].EconomicsData.MaxGasLimitPerBlock(0) - 1,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
		Data:     []byte(upgradeTxData),
	}, nodes)

	nonce, round = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	require.Equal(t, []byte{24}, query(t, nodes[0], scAddress, "getUltimateAnswer"))

	// Now upgrade as Alice, should work
	addTxToPool(&transaction.Transaction{
		Nonce:    1,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  alice,
		GasPrice: nodes[0].EconomicsData.GetMinGasPrice(),
		GasLimit: nodes[0].EconomicsData.MaxGasLimitPerBlock(0) - 1,
		ChainID:  integrationTests.ChainID,
		Version:  integrationTests.MinTransactionVersion,
		Data:     []byte(upgradeTxData),
	}, nodes)

	_, _ = integrationTests.WaitOperationToBeDone(t, nodes, 1, nonce, round, idxProposers)
	require.Equal(t, []byte{42}, query(t, nodes[0], scAddress, "getUltimateAnswer"))
}

func addTxToPool(tx *transaction.Transaction, nodes []*integrationTests.TestProcessorNode) {
	txHash, _ := core.CalculateHash(integrationTests.TestMarshalizer, integrationTests.TestHasher, tx)
	sourceShard := nodes[0].ShardCoordinator.ComputeId(tx.SndAddr)

	for _, node := range nodes {
		if node.ShardCoordinator.SelfId() != sourceShard {
			continue
		}

		cacheIdentifier := process.ShardCacherIdentifier(sourceShard, sourceShard)
		node.DataPool.Transactions().AddData(txHash, tx, tx.Size(), cacheIdentifier)
	}
}

func query(t *testing.T, node *integrationTests.TestProcessorNode, scAddress []byte, function string) []byte {
	scQuery := node.SCQueryService
	vmOutput, err := scQuery.ExecuteQuery(&process.SCQuery{
		ScAddress: scAddress,
		FuncName:  function,
		Arguments: [][]byte{},
	})

	require.Nil(t, err)
	require.Equal(t, vmcommon.Ok, vmOutput.ReturnCode)
	return vmOutput.ReturnData[0]
}
