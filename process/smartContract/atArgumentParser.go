package smartContract

import "math/big"

type atArgumentParser struct {
}

func NewAtArgumentParser() *atArgumentParser {
	return &atArgumentParser{}
}

func (at *atArgumentParser) CreateArguments(data []byte) []*big.Int {
	args := make([]*big.Int, 0)
	return args
}
