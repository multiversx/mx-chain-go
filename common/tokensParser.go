package common

import (
	"math/big"
	"strings"

	"github.com/multiversx/mx-chain-go/common/tokens"
)

const (
	// esdtTickerNumRandChars represents the number of hex-encoded random characters sequence of a ticker
	esdtTickerNumRandChars = 6

	// separatorChar represents the character that separated the token ticker by the random sequence
	separatorChar = "-"
)

// TODO: move this to core

// ExtractTokenIDAndNonceFromTokenStorageKey will parse the token's storage key and extract the identifier and the nonce
func ExtractTokenIDAndNonceFromTokenStorageKey(tokenKey []byte) ([]byte, uint64) {
	// ALC-1q2w3e for fungible
	// ALC-2w3e4rX for non fungible
	token := string(tokenKey)

	// filtering by the index of first occurrence is faster than splitting
	indexOfFirstHyphen := strings.Index(token, separatorChar)
	if indexOfFirstHyphen < 0 {
		return tokenKey, 0
	}

	indexOfTokenHyphen := getIndexOfTokenHyphen(token, indexOfFirstHyphen)
	tokenTicker := token[:indexOfTokenHyphen]
	randomSequencePlusNonce := token[indexOfTokenHyphen+1:]

	tokenTickerLen := len(tokenTicker)

	areTickerAndRandomSequenceInvalid := !tokens.IsTokenTickerLenCorrect(tokenTickerLen) || len(randomSequencePlusNonce) == 0
	if areTickerAndRandomSequenceInvalid {
		return tokenKey, 0
	}

	if len(randomSequencePlusNonce) < esdtTickerNumRandChars+1 {
		return tokenKey, 0
	}

	// ALC-1q2w3eX - X is the nonce
	nonceStr := randomSequencePlusNonce[esdtTickerNumRandChars:]
	nonceBigInt := big.NewInt(0).SetBytes([]byte(nonceStr))

	numCharsSinceNonce := len(token) - len(nonceStr)
	tokenID := token[:numCharsSinceNonce]

	return []byte(tokenID), nonceBigInt.Uint64()
}

func getIndexOfTokenHyphen(token string, indexOfFirstHyphen int) int {
	if !isValidPrefixedToken(token) {
		return indexOfFirstHyphen
	}

	indexOfSecondHyphen := strings.Index(token[indexOfFirstHyphen+1:], separatorChar)
	return indexOfSecondHyphen + indexOfFirstHyphen + 1
}

func isValidPrefixedToken(token string) bool {
	tokenSplit := strings.Split(token, separatorChar)
	if len(tokenSplit) < 3 {
		return false
	}

	prefix := tokenSplit[0]
	if !tokens.IsValidTokenPrefix(prefix) {
		return false
	}

	tokenTicker := tokenSplit[1]
	if !tokens.IsTickerValid(tokenTicker) {
		return false
	}

	tokenRandSeq := tokenSplit[2]
	return len(tokenRandSeq) >= esdtTickerNumRandChars
}
