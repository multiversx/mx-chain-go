package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractTokenIDAndNonceFromTokenStorageKey(t *testing.T) {
	t.Parallel()

	t.Run("regular esdt, should work", func(t *testing.T) {
		t.Parallel()

		id := "ALC-1q2w3e"
		tokenName, nonce := ExtractTokenIDAndNonceFromTokenStorageKey([]byte(id))
		require.Equal(t, uint64(0), nonce)
		require.Equal(t, []byte(id), tokenName)
	})

	t.Run("non fungible, should work", func(t *testing.T) {
		t.Parallel()

		nonceBI := big.NewInt(37)
		id := "ALC-1q2w3e"
		tokenKey := id + string(nonceBI.Bytes())
		tokenName, nonce := ExtractTokenIDAndNonceFromTokenStorageKey([]byte(tokenKey))
		require.Equal(t, nonceBI.Uint64(), nonce)
		require.Equal(t, []byte(id), tokenName)
	})
}
