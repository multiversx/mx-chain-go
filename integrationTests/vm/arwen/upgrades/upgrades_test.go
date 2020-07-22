package upgrades

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/integrationTests/vm/arwen"
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
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_StorageCannotBeModifiedByNonOwner(t *testing.T) {
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
	require.Nil(t, err)
	require.Equal(t, uint64(2), context.QuerySCInt("get", [][]byte{}))
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

	context.GasLimit = 21600000
	err := context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Equal(t, fmt.Errorf("execution failed"), err)

	context.GasLimit = 21700000
	err = context.DeploySC(delegationWasmPath, delegationInitParams)
	require.Nil(t, err)

	err = context.UpgradeSC(delegationWasmPath, delegationUpgradeParams)
	require.Nil(t, err)
}
