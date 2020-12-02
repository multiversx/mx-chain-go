package parsers

import (
	"encoding/hex"
	"strings"
)

func tokenize(data string) ([]string, error) {
	tokens := strings.Split(data, atSeparator)

	if len(tokens) == 0 || len(tokens[0]) == 0 {
		return nil, ErrTokenizeFailed
	}

	return tokens, nil
}

func decodeToken(token string) ([]byte, error) {
	decoded, err := hex.DecodeString(token)
	if err != nil {
		return nil, ErrTokenizeFailed
	}

	return decoded, nil
}

func trimLeadingSeparatorChar(data string) string {
	if len(data) > 0 && data[0] == atSeparatorChar {
		data = data[1:]
	}

	return data
}

func requireNumTokensIsEven(tokens []string) error {
	if len(tokens)%2 == 0 {
		return nil
	}

	return ErrInvalidDataString
}
