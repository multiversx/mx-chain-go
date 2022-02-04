package common

import (
	"math/big"
	"strings"
)

// esdtTickerNumRandChars represents the number of hex-encoded random characters sequence of a ticker
const esdtTickerNumRandChars = 6

// TODO: move this to core

// ExtractTokenIDAndNonceFromTokenKey will parse the token's storage key and extract the identifier and the nonce
func ExtractTokenIDAndNonceFromTokenStorageKey(tokenKey []byte) ([]byte, uint64) {
	// ALC-1q2w3e for fungible
	// ALC-2w3e4rX for non fungible
	token := string(tokenKey)
	splitToken := strings.Split(token, "-")
	if len(splitToken) < 2 {
		return tokenKey, 0
	}

	if len(splitToken[1]) < esdtTickerNumRandChars+ 1 {
		return tokenKey, 0
	}

	// ALC-1q2w3eX - X is the nonce
	nonceStr := splitToken[1][esdtTickerNumRandChars:]
	nonceBigInt := big.NewInt(0).SetBytes([]byte(nonceStr))

	numCharsSinceNonce := len(token) - len(nonceStr)
	tokenID := token[:numCharsSinceNonce]

	return []byte(tokenID), nonceBigInt.Uint64()
}
