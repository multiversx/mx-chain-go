package upgrades

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/integrationTests"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm"
	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory"
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
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
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
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
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

	childUpgradedCode := arwen.GetSCCode("../testdata/hello-v2/output/answer.wasm")
	err = context.ExecuteSC(owner, "upgradeChild@"+childUpgradedCode)
	require.Nil(t, err)

	context.ScAddress = childAddress
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
	parentAddress = context.ScAddress
}

func TestUpgrades_UpgradeDelegationContract(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	delegationWasmPath := "../testdata/delegation/delegation.wasm"
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"
	delegationUpgradeParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"

	context.ScCodeMetadata.Upgradeable = true
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
	delegationInitParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"
	delegationUpgradeParams := "0000000000000000000000000000000000000000000000000000000000000000@03E8@00@030D40@030D40"

	context.GasLimit = 21700000
	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(t, err)

	context.Owner = context.Alice
	err = context.UpgradeSC(delegationWasmPath, delegationUpgradeParams)
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
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
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
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
	require.Equal(t, process.ErrUpgradeNotAllowed, err)
	require.Equal(t, uint64(2), context.QuerySCInt("get", [][]byte{}))
}

func TestUpgrades_HelloTrialAndError(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	network := integrationTests.NewOneNodeNetwork()
	defer network.Stop()

	alice := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bob := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	network.Mint(alice, big.NewInt(10000000000000))
	network.Mint(bob, big.NewInt(10000000000000))

	network.GoToRoundOne()

	deployTxData := fmt.Sprintf("%s@%s@0100", arwen.GetSCCode("../testdata/hello-v1/output/answer.wasm"), hex.EncodeToString(factory.ArwenVirtualMachine))
	upgradeTxData := fmt.Sprintf("upgradeContract@%s@0100", arwen.GetSCCode("../testdata/hello-v2/output/answer.wasm"))

	// Deploy the smart contract. Alice is the owner
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress(),
		SndAddr:  alice,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(deployTxData),
	})

	scAddress, _ := network.Node.BlockchainHook.NewAddress(alice, 0, factory.ArwenVirtualMachine)
	network.Continue(t, 1)
	require.Equal(t, []byte{24}, query(t, network.Node, scAddress, "getUltimateAnswer"))

	// Upgrade as Bob - upgrade should fail, since Alice is the owner
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  bob,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(upgradeTxData),
	})

	network.Continue(t, 1)
	require.Equal(t, []byte{24}, query(t, network.Node, scAddress, "getUltimateAnswer"))

	// Now upgrade as Alice, should work
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    1,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  alice,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(upgradeTxData),
	})

	network.Continue(t, 1)
	require.Equal(t, []byte{42}, query(t, network.Node, scAddress, "getUltimateAnswer"))
}

func TestUpgrades_CounterTrialAndError(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	network := integrationTests.NewOneNodeNetwork()
	defer network.Stop()

	alice := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bob := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	network.Mint(alice, big.NewInt(10000000000000))
	network.Mint(bob, big.NewInt(10000000000000))

	network.GoToRoundOne()

	deployTxData := fmt.Sprintf("%s@%s@0100", arwen.GetSCCode("../testdata/counter/output/counter.wasm"), hex.EncodeToString(factory.ArwenVirtualMachine))
	upgradeTxData := fmt.Sprintf("upgradeContract@%s@0100", arwen.GetSCCode("../testdata/counter/output/counter.wasm"))

	// Deploy the smart contract. Alice is the owner
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  vm.CreateEmptyAddress(),
		SndAddr:  alice,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(deployTxData),
	})

	scAddress, _ := network.Node.BlockchainHook.NewAddress(alice, 0, factory.ArwenVirtualMachine)
	network.Continue(t, 1)
	require.Equal(t, []byte{1}, query(t, network.Node, scAddress, "get"))

	// Increment the counter (could be either Bob or Alice)
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    1,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  alice,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte("increment"),
	})

	network.Continue(t, 1)
	require.Equal(t, []byte{2}, query(t, network.Node, scAddress, "get"))

	// Upgrade as Bob - upgrade should fail, since Alice is the owner (counter.init() not executed, state not reset)
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    0,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  bob,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(upgradeTxData),
	})

	network.Continue(t, 1)
	require.Equal(t, []byte{2}, query(t, network.Node, scAddress, "get"))

	// Now upgrade as Alice, should work (state is reset by counter.init())
	network.AddTxToPool(&transaction.Transaction{
		Nonce:    2,
		Value:    big.NewInt(0),
		RcvAddr:  scAddress,
		SndAddr:  alice,
		GasPrice: network.GetMinGasPrice(),
		GasLimit: network.MaxGasLimitPerBlock(),
		Data:     []byte(upgradeTxData),
	})

	network.Continue(t, 1)
	require.Equal(t, []byte{1}, query(t, network.Node, scAddress, "get"))
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
