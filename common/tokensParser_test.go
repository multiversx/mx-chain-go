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

		checkTickerAndNonceExtraction(t, "ALC-1q2w3e", "ALC-1q2w3e", 0)
	})

	t.Run("non fungible, should work", func(t *testing.T) {
		t.Parallel()

		nonceBI := big.NewInt(37)
		id := "ALC-1q2w3e"
		tokenKey := id + string(nonceBI.Bytes())

		checkTickerAndNonceExtraction(t, tokenKey, id, nonceBI.Uint64())
	})

	t.Run("malformed keys, should be treated correctly", func(t *testing.T) {
		t.Parallel()

		checkTickerAndNonceExtraction(t, "INVALID_KEY", "INVALID_KEY", 0)

		checkTickerAndNonceExtraction(t, "--------", "--------", 0)

		checkTickerAndNonceExtraction(t, "a--------", "a--------", 0)

		checkTickerAndNonceExtraction(t, "abcdef", "abcdef", 0)

		checkTickerAndNonceExtraction(t, "a-b-c-d-e-f", "a-b-c-d-e-f", 0)

		checkTickerAndNonceExtraction(t, "abc-def", "abc-def", 0)

		checkTickerAndNonceExtraction(t, "abcdef-", "abcdef-", 0)

		checkTickerAndNonceExtraction(t, "abcdefabcdefabcdef-", "abcdefabcdefabcdef-", 0)

		checkTickerAndNonceExtraction(t, "ab-cdefabcdefabcdef", "ab-cdefabcdefabcdef", 0)
	})

	t.Run("test for edge case when the nonce is the hyphen ascii char", func(t *testing.T) {
		t.Parallel()

		// "-" represents nonce 45 and should not be treated as a separator
		checkTickerAndNonceExtraction(t, "EGLDMEXF-8aa8b6-", "EGLDMEXF-8aa8b6", 45)
	})

	t.Run("prefixed tokens", func(t *testing.T) {
		t.Parallel()

		checkTickerAndNonceExtraction(t, "pref-ALC-1q2w3e", "pref-ALC-1q2w3e", 0)
		checkTickerAndNonceExtraction(t, "pf1-ALC-1q2w3e", "pf1-ALC-1q2w3e", 0)

		checkTickerAndNonceExtraction(t, "sv1-TKN-1q2w3e4", "sv1-TKN-1q2w3e", 52)
		checkTickerAndNonceExtraction(t, "sv1-TKN-1q2w3e-", "sv1-TKN-1q2w3e", 45)
	})
}

func checkTickerAndNonceExtraction(t *testing.T, input string, expectedTicker string, expectedNonce uint64) {
	tokenName, nonce := ExtractTokenIDAndNonceFromTokenStorageKey([]byte(input))
	require.Equal(t, expectedNonce, nonce)
	require.Equal(t, []byte(expectedTicker), tokenName)
}
