package smartContract

import "math/big"

type atArgumentParser struct {
}

func NewAtArgumentParser() *atArgumentParser {
	return &atArgumentParser{}
}

func (at *atArgumentParser) CreateArguments(data []byte) ([]*big.Int, error) {
	args := make([]*big.Int, 0)
	return args, nil
}

func (at *atArgumentParser) GetCodeFromData(data []byte) ([]byte, error) {
	return []byte(""), nil
}

func (at *atArgumentParser) GetFunctionFromData(data []byte) ([]byte, error) {
	return []byte(""), nil
}
