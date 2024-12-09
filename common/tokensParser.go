package common

import (
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-core-go/data/esdt"
)

const (
	// esdtTickerNumRandChars represents the number of hex-encoded random characters sequence of a ticker
	esdtTickerNumRandChars = 6

	// separatorChar represents the character that separated the token ticker by the random sequence
	separatorChar = "-"
)

type tokenData struct {
	hasPrefix               bool
	ticker                  string
	randomSequencePlusNonce string
}

// TODO: move this to core

// ExtractTokenIDAndNonceFromTokenStorageKey will parse the token's storage key and extract the identifier and the nonce
// It also returns true if it has a valid prefix.
func ExtractTokenIDAndNonceFromTokenStorageKey(tokenKey []byte) ([]byte, bool, uint64) {
	// ALC-1q2w3e for fungible
	// ALC-2w3e4rX for non fungible
	token := string(tokenKey)

	// filtering by the index of first occurrence is faster than splitting
	indexOfFirstHyphen := strings.Index(token, separatorChar)
	if indexOfFirstHyphen < 0 {
		return tokenKey, false, 0
	}

	tknData := getTokenData(token, indexOfFirstHyphen)
	if !isTokenDataValid(tknData) {
		return tokenKey, tknData.hasPrefix, 0
	}

	// ALC-1q2w3eX - X is the nonce
	nonceStr := tknData.randomSequencePlusNonce[esdtTickerNumRandChars:]
	nonceBigInt := big.NewInt(0).SetBytes([]byte(nonceStr))

	numCharsSinceNonce := len(token) - len(nonceStr)
	tokenID := token[:numCharsSinceNonce]

	return []byte(tokenID), tknData.hasPrefix, nonceBigInt.Uint64()
}

func getTokenData(token string, indexOfFirstHyphen int) *tokenData {
	if !isValidPrefixedToken(token) {
		return &tokenData{
			ticker:                  token[:indexOfFirstHyphen],
			randomSequencePlusNonce: token[indexOfFirstHyphen+1:],
			hasPrefix:               false,
		}
	}

	indexOfSecondHyphen := strings.Index(token[indexOfFirstHyphen+1:], separatorChar)
	indexOfTokenHyphen := indexOfSecondHyphen + indexOfFirstHyphen + 1
	return &tokenData{
		ticker:                  token[indexOfFirstHyphen+1 : indexOfTokenHyphen],
		randomSequencePlusNonce: token[indexOfTokenHyphen+1:],
		hasPrefix:               true,
	}
}

func isTokenDataValid(tknData *tokenData) bool {
	tokenTickerLen := len(tknData.ticker)
	isTickerValid := esdt.IsTickerValid(tknData.ticker) && esdt.IsTokenTickerLenCorrect(tokenTickerLen)
	isRandomSequencePlusNonceValid := len(tknData.randomSequencePlusNonce) >= esdtTickerNumRandChars+1

	return isTickerValid && isRandomSequencePlusNonceValid
}

func isValidPrefixedToken(token string) bool {
	tokenSplit := strings.Split(token, separatorChar)
	if len(tokenSplit) < 3 {
		return false
	}

	prefix := tokenSplit[0]
	if !esdt.IsValidTokenPrefix(prefix) {
		return false
	}

	tokenTicker := tokenSplit[1]
	if !esdt.IsTickerValid(tokenTicker) {
		return false
	}

	tokenRandSeq := tokenSplit[2]
	return len(tokenRandSeq) >= esdtTickerNumRandChars
}
