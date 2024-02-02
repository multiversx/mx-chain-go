package metachain

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPrettyValue(t *testing.T) {
	require.Equal(t, "1234.0", getPrettyValue(big.NewInt(1234), big.NewInt(1)))
	require.Equal(t, "123.4", getPrettyValue(big.NewInt(1234), big.NewInt(10)))
	require.Equal(t, "12.34", getPrettyValue(big.NewInt(1234), big.NewInt(100)))
	require.Equal(t, "1.234", getPrettyValue(big.NewInt(1234), big.NewInt(1000)))
	require.Equal(t, "0.1234", getPrettyValue(big.NewInt(1234), big.NewInt(10000)))
	require.Equal(t, "0.01234", getPrettyValue(big.NewInt(1234), big.NewInt(100000)))
	require.Equal(t, "0.00123", getPrettyValue(big.NewInt(1234), big.NewInt(1000000)))
	require.Equal(t, "0.00012", getPrettyValue(big.NewInt(1234), big.NewInt(10000000)))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(1234), big.NewInt(100000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1234), big.NewInt(1000000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1234), big.NewInt(10000000000)))

	require.Equal(t, "1.0", getPrettyValue(big.NewInt(1), big.NewInt(1)))
	require.Equal(t, "0.1", getPrettyValue(big.NewInt(1), big.NewInt(10)))
	require.Equal(t, "0.01", getPrettyValue(big.NewInt(1), big.NewInt(100)))
	require.Equal(t, "0.001", getPrettyValue(big.NewInt(1), big.NewInt(1000)))
	require.Equal(t, "0.0001", getPrettyValue(big.NewInt(1), big.NewInt(10000)))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(1), big.NewInt(100000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1), big.NewInt(1000000)))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1), big.NewInt(10000000)))

	oneEGLD := big.NewInt(1000000000000000000)
	denominationEGLD := big.NewInt(int64(math.Pow10(18)))

	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(0), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(oneEGLD, denominationEGLD))
	require.Equal(t, "1.10000", getPrettyValue(big.NewInt(1100000000000000000), denominationEGLD))
	require.Equal(t, "1.10000", getPrettyValue(big.NewInt(1100000000000000001), denominationEGLD))
	require.Equal(t, "1.11000", getPrettyValue(big.NewInt(1110000000000000001), denominationEGLD))
	require.Equal(t, "0.11100", getPrettyValue(big.NewInt(111000000000000001), denominationEGLD))
	require.Equal(t, "0.01110", getPrettyValue(big.NewInt(11100000000000001), denominationEGLD))
	require.Equal(t, "0.00111", getPrettyValue(big.NewInt(1110000000000001), denominationEGLD))
	require.Equal(t, "0.00011", getPrettyValue(big.NewInt(111000000000001), denominationEGLD))
	require.Equal(t, "0.00001", getPrettyValue(big.NewInt(11100000000001), denominationEGLD))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(1110000000001), denominationEGLD))
	require.Equal(t, "0.00000", getPrettyValue(big.NewInt(111000000001), denominationEGLD))

	require.Equal(t, "2.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(2)), denominationEGLD))
	require.Equal(t, "20.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(20)), denominationEGLD))
	require.Equal(t, "2000000.00000", getPrettyValue(big.NewInt(0).Mul(oneEGLD, big.NewInt(2000000)), denominationEGLD))

	require.Equal(t, "3.22220", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000000000)), denominationEGLD))
	require.Equal(t, "1.22222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000000000)), denominationEGLD))
	require.Equal(t, "1.02222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(22222000000000000)), denominationEGLD))
	require.Equal(t, "1.00222", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000000)), denominationEGLD))
	require.Equal(t, "1.00022", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000000)), denominationEGLD))
	require.Equal(t, "1.00002", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(22222000000000)), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(2222200000000)), denominationEGLD))
	require.Equal(t, "1.00000", getPrettyValue(big.NewInt(0).Add(oneEGLD, big.NewInt(222220000000)), denominationEGLD))
}
