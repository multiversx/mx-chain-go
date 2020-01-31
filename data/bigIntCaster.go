package data

import (
	"fmt"
	"math/big"
)

type BigIntCaster struct{}

func (c *BigIntCaster) Equal(a, b *big.Int) bool {
	if a == nil {
		return b == nil
	}
	return a.Cmp(b) == 0
}

func (c *BigIntCaster) Size(a *big.Int) int {
	return len(a.Bytes()) + 1
}

func (c *BigIntCaster) MarshalTo(a *big.Int, buf []byte) (int, error) {
	bytes := a.Bytes()
	if len(buf) <= len(bytes) {
		return 0, ErrInvalidValue
	}
	if a == nil {
		buf[0] = 0
		buf[1] = 0
		return 2, nil
	}
	copy(buf[1:], bytes)
	if a.Sign() < 0 {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	return len(bytes) + 1, nil
}

func (c *BigIntCaster) Unmarshal(buf []byte) (*big.Int, error) {
	if len(buf) == 0 {
		return big.NewInt(0), fmt.Errorf("bad input")
	}
	ret := new(big.Int).SetBytes(buf[1:])
	switch buf[0] {
	case 0:

	case 1:
		ret = ret.Neg(ret)
	default:
		return big.NewInt(0), fmt.Errorf("invalid sign byte %x", buf[0])
	}

	return ret, nil
}

func (c *BigIntCaster) NewPopulated() *big.Int {
	return big.NewInt(0)
}
