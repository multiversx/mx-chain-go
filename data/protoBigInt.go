package data

import (
	"math/big"
)

type ProtoBigInt struct {
	big.Int
}

// NewProtoBigInt allocates and returns a new ProtoBigInt set to x.
func NewProtoBigInt(x int64) *ProtoBigInt {
	t := ProtoBigInt{}
	t.SetInt64(x)
	return &t
}

// NewProtoBigInt allocates and returns a new ProtoBigInt set to bi.
func NewProtoBigIntFromBigInt(bi *big.Int) *ProtoBigInt {
	if bi == nil {
		return nil
	}
	t := ProtoBigInt{*bi}
	return &t
}

func writeToSlice(sign int, abs, data []byte) {

	if sign == -1 {
		data[0] = 1
	} else {
		data[0] = 0
	}

	copy(data[1:], abs)
}

func (t ProtoBigInt) Marshal() ([]byte, error) {
	tmp := t.Bytes()
	data := make([]byte, len(tmp)+1)
	writeToSlice(t.Sign(), tmp, data)
	return data, nil

}

func (t *ProtoBigInt) Text(base int) string {
	if t == nil {
		var nbi *big.Int
		return nbi.Text(base)
	}
	return t.Int.Text(base)
}

func (t *ProtoBigInt) String() string {
	return t.Text(10)
}

func (t *ProtoBigInt) MarshalTo(data []byte) (n int, err error) {
	tmp := t.Bytes()
	if len(data) <= len(tmp) {
		return 0, ErrInvalidValue
	}

	writeToSlice(t.Sign(), tmp, data)
	return len(tmp) + 1, nil
}

func (t *ProtoBigInt) Unmarshal(data []byte) error {
	if len(data) < 1 {
		return ErrInvalidValue
	}
	t.SetBytes(data[1:])
	if data[0] == 0 {
		// Maybe revize this
	} else if data[0] == 1 {
		t.Mul(&t.Int, big.NewInt(-1))
	} else {
		t.SetInt64(0)
		return ErrInvalidValue
	}
	return nil

}
func (t *ProtoBigInt) Size() int {
	tmp := t.Bytes()
	return len(tmp) + 1 /*For sign*/
}

func (t ProtoBigInt) MarshalJSON() ([]byte, error) {
	str := t.String()
	ret := make([]byte, len(str)+2)
	ret[0] = '"'
	copy(ret[1:], []byte(str))
	ret[len(ret)-1] = '"'
	return ret, nil
}

func (t *ProtoBigInt) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return ErrInvalidValue
	}
	if _, ok := t.SetString(string(data[1:len(data)-1]), 10); !ok {
		return ErrInvalidValue
	}
	return nil

}

// only required if the compare option is set
func (t ProtoBigInt) Compare(other ProtoBigInt) int {
	return t.Cmp(&other.Int)
}

// only required if the equal option is set
func (t ProtoBigInt) Equal(other ProtoBigInt) bool {
	return t.Cmp(&other.Int) == 0
}
