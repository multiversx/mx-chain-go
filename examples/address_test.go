package examples

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/display"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHexAddressToBech32Address(t *testing.T) {
	t.Parallel()

	hexEncodedAddress := "af006ece83473104ea91f7ff5605c4c1742f7214a1f46be299e30ee2e8707169"

	hexEncodedAddressBytes, err := hex.DecodeString(hexEncodedAddress)
	require.NoError(t, err)

	bech32Address := addressEncoder.Encode(hexEncodedAddressBytes)
	require.Equal(t, "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49", bech32Address)
}

func TestBech32AddressToHexAddress(t *testing.T) {
	t.Parallel()

	bech32Address := "erd14uqxan5rgucsf6537ll4vpwyc96z7us5586xhc5euv8w96rsw95sfl6a49"

	bech32AddressBytes, err := addressEncoder.Decode(bech32Address)
	require.NoError(t, err)

	hexEncodedAddress := hex.EncodeToString(bech32AddressBytes)
	require.Equal(t, "af006ece83473104ea91f7ff5605c4c1742f7214a1f46be299e30ee2e8707169", hexEncodedAddress)
}

func TestShardOfAddress(t *testing.T) {
	t.Parallel()

	// the shard of an address depends on the number of shards in the chain. The same address does not necessarily
	// belong to the same shard in a chain with a different number of shards.

	numberOfShards := uint32(3)
	shardCoordinator, err := sharding.NewMultiShardCoordinator(numberOfShards, 0)
	require.NoError(t, err)

	require.Equal(t, uint32(0), computeShardID(t, "erd1gn0y4l4rgkf2e7dg74u3nnugr7uycw5jwa44tlnqg2kxa37dr2kq60xwvg", shardCoordinator))
	require.Equal(t, uint32(1), computeShardID(t, "erd1x23lzn8483xs2su4fak0r0dqx6w38enpmmqf2yrkylwq7mfnvyhsxqw57y", shardCoordinator))
	require.Equal(t, uint32(2), computeShardID(t, "erd1zwkdd3k023llluhkd0963kdtfjh0xfgh8ngfwt2qj9da0l79qgpqvqluqd", shardCoordinator))
	require.Equal(t, core.MetachainShardId, computeShardID(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u", shardCoordinator))
}

func computeShardID(t *testing.T, addressBech32 string, shardCoordinator sharding.Coordinator) uint32 {
	addressBytes, err := addressEncoder.Decode(addressBech32)
	require.NoError(t, err)

	return shardCoordinator.ComputeId(addressBytes)
}

func TestSystemSCsAddressesAndSpecialAddresses(t *testing.T) {
	contractDeployScAdress := addressEncoder.Encode(make([]byte, addressEncoder.Len()))
	stakingScAddress := addressEncoder.Encode(vm.StakingSCAddress)
	validatorScAddress := addressEncoder.Encode(vm.ValidatorSCAddress)
	esdtScAddress := addressEncoder.Encode(vm.ESDTSCAddress)
	governanceScAddress := addressEncoder.Encode(vm.GovernanceSCAddress)
	jailingAddress := addressEncoder.Encode(vm.JailingAddress)
	endOfEpochAddress := addressEncoder.Encode(vm.EndOfEpochAddress)
	delegationManagerScAddress := addressEncoder.Encode(vm.DelegationManagerSCAddress)
	firstDelegationScAddress := addressEncoder.Encode(vm.FirstDelegationSCAddress)

	header := []string{"Smart contract/Special address", "Address"}
	lines := []*display.LineData{
		display.NewLineData(false, []string{"Contract deploy", contractDeployScAdress}),
		display.NewLineData(false, []string{"Staking", stakingScAddress}),
		display.NewLineData(false, []string{"Validator", validatorScAddress}),
		display.NewLineData(false, []string{"ESDT", esdtScAddress}),
		display.NewLineData(false, []string{"Governance", governanceScAddress}),
		display.NewLineData(false, []string{"Jailing address", jailingAddress}),
		display.NewLineData(false, []string{"End of epoch address", endOfEpochAddress}),
		display.NewLineData(false, []string{"Delegation manager", delegationManagerScAddress}),
		display.NewLineData(false, []string{"First delegation", firstDelegationScAddress}),
	}

	table, _ := display.CreateTableString(header, lines)
	fmt.Println(table)

	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqllls0lczs7", stakingScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqplllst77y4l", validatorScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqzllls8a5w6u", esdtScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqrlllsrujgla", governanceScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqrlllllllllllllllllllllllllsn60f0k", jailingAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqlllllllllllllllllllllllllllsr9gav8", endOfEpochAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqylllslmq6y6", delegationManagerScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqpqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq0llllsqkarq6", firstDelegationScAddress)
	assert.Equal(t, "erd1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq6gq4hu", contractDeployScAdress)
}
