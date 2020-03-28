package upgrades

import (
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
	err := context.DeploySC("../testdata/hello-v1/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2")

	err = context.UpgradeSC("../testdata/hello-v2/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v3")

	err = context.UpgradeSC("../testdata/hello-v3/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, "forty-two", context.QuerySCString("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_HelloDoesNotUpgradeWhenNotUpgradeable(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = false
	err := context.DeploySC("../testdata/hello-v1/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2 will not be performed")

	err = context.UpgradeSC("../testdata/hello-v2/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}

func TestUpgrades_HelloUpgradesToNotUpgradeable(t *testing.T) {
	context := arwen.SetupTestContext(t)
	defer context.Close()

	fmt.Println("Deploy v1")

	context.ScCodeMetadata.Upgradeable = true
	err := context.DeploySC("../testdata/hello-v1/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(24), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v2, becomes not upgradeable")

	context.ScCodeMetadata.Upgradeable = false
	err = context.UpgradeSC("../testdata/hello-v2/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))

	fmt.Println("Upgrade to v3")

	err = context.UpgradeSC("../testdata/hello-v3/answer.wasm", "")
	require.Nil(t, err)
	require.Equal(t, uint64(42), context.QuerySCInt("getUltimateAnswer", [][]byte{}))
}
