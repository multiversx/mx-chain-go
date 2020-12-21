package examples

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/sharding"
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
