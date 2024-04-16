package common

import (
	"math/big"
	"strings"
)

const (
	// esdtTickerNumRandChars represents the number of hex-encoded random characters sequence of a ticker
	esdtTickerNumRandChars = 6

	// separatorChar represents the character that separated the token ticker by the random sequence
	separatorChar = "-"

	// minLengthForTickerName represents the minimum number of characters a token's ticker can have
	minLengthForTickerName = 3

	// maxLengthForTickerName represents the maximum number of characters a token's ticker can have
	maxLengthForTickerName = 10

	esdtMinPrefixLen = 4
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

	token = getUnPrefixedToken(token)
	tokenTicker := token[:indexOfFirstHyphen]
	randomSequencePlusNonce := token[indexOfFirstHyphen+1:]

	tokenTickerLen := len(tokenTicker)

	areTickerAndRandomSequenceInvalid := !isTokenTickerLenCorrect(tokenTickerLen) || len(randomSequencePlusNonce) == 0
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

func getUnPrefixedToken(token string) string {
	tokenSplit := strings.Split(token, separatorChar)
	if len(tokenSplit) < 2 {
		return token
	}

	prefix := tokenSplit[0]
	if !isValidPrefix(prefix) {
		return token
	}

	tokenTicker := tokenSplit[1]
	tokenRandSeq := tokenSplit[2]
	if !(isTokenTickerLenCorrect(len(tokenTicker)) && len(tokenRandSeq) == esdtTickerNumRandChars) {
		return token
	}

	indexOfFirstHyphen := strings.Index(token, separatorChar)
	return token[indexOfFirstHyphen+1:]
}

func isValidPrefix(prefix string) bool {
	if len(prefix) > esdtMinPrefixLen {
		return false
	}

	for _, ch := range prefix {
		isLowerCaseCharacter := ch >= 'a' && ch <= 'z'
		isNumber := ch >= '0' && ch <= '9'
		isAllowedPrefix := isLowerCaseCharacter || isNumber
		if !isAllowedPrefix {
			return false
		}
	}

	return true
}

func isTokenTickerLenCorrect(tokenTickerLen int) bool {
	return !(tokenTickerLen == 0 ||
		tokenTickerLen < minLengthForTickerName ||
		tokenTickerLen > maxLengthForTickerName)
}
