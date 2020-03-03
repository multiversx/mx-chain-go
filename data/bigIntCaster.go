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
	if a == nil {
		return 1
	}
	if size := len(a.Bytes()); size > 0 {
		return size + 1
	}
	return 2
}

func (c *BigIntCaster) MarshalTo(a *big.Int, buf []byte) (int, error) {
	if a == nil {
		buf[0] = 0
		return 1, nil
	}
	bytes := a.Bytes()
	if len(buf) <= len(bytes) {
		return 0, ErrInvalidValue
	}
	copy(buf[1:], bytes)
	if a.Sign() < 0 {
		buf[0] = 1
	} else {
		buf[0] = 0
	}
	bsize := len(bytes)
	if bsize > 0 {
		return bsize + 1, nil
	}
	return 2, nil
}

func (c *BigIntCaster) Unmarshal(buf []byte) (*big.Int, error) {
	switch len(buf) {
	case 0:
		return nil, fmt.Errorf("bad input")
	case 1:
		return nil, nil
	case 2:
		if buf[1] == 0 {
			return big.NewInt(0), nil
		}

	}
	ret := new(big.Int).SetBytes(buf[1:])
	switch buf[0] {
	case 0:

	case 1:
		ret = ret.Neg(ret)
	default:
		return nil, fmt.Errorf("invalid sign byte %x", buf[0])
	}

	return ret, nil
}

func (c *BigIntCaster) NewPopulated() *big.Int {
	return big.NewInt(0)
}
