package data

import (
	"math/big"
)

type ProtoBigInt struct {
	i big.Int
}

// NewProtoBigInt allocates and returns a new ProtoBigInt set to x.
func NewProtoBigInt(x int64) *ProtoBigInt {
	t := ProtoBigInt{}
	t.i.SetInt64(x)
	return &t
}

// NewProtoBigInt allocates and returns a new ProtoBigInt set to bi.
func NewProtoBigIntFromBigInt(bi *big.Int) *ProtoBigInt {
	t := ProtoBigInt{}
	if bi != nil {
		t.i.Set(bi)
	}
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

// Marshal used in gogo protobuf conversion
func (t ProtoBigInt) Marshal() ([]byte, error) {
	tmp := t.i.Bytes()
	data := make([]byte, len(tmp)+1)
	writeToSlice(t.i.Sign(), tmp, data)
	return data, nil
}

// MarshalTo used in gogo protobuf conversion
func (t *ProtoBigInt) MarshalTo(data []byte) (n int, err error) {
	tmp := t.i.Bytes()
	if len(data) <= len(tmp) {
		return 0, ErrInvalidValue
	}

	writeToSlice(t.i.Sign(), tmp, data)
	return len(tmp) + 1, nil
}

// Unmarshal used in gogo protobuf conversion
func (t *ProtoBigInt) Unmarshal(data []byte) error {
	if len(data) < 1 || data[0] > 1 {
		return ErrInvalidValue
	}
	t.i.SetBytes(data[1:])
	if data[0] == 1 {
		t.i.Neg(&t.i)
	}
	return nil

}

// Size get the protobuf size of t
func (t *ProtoBigInt) Size() int {
	tmp := t.i.Bytes()
	return len(tmp) + 1 /*For sign*/
}

// Text safe conversion of t to its base string representation
func (t *ProtoBigInt) Text(base int) string {
	if t == nil {
		var nbi *big.Int
		return nbi.Text(base)
	}
	return t.i.Text(base)
}

// Text safe conversion of t to its base 10 string representation
func (t *ProtoBigInt) String() string {
	return t.Text(10)
}

// MarshalJSON convert t to its JSON representation
func (t ProtoBigInt) MarshalJSON() ([]byte, error) {
	str := t.String()
	ret := make([]byte, len(str)+2)
	ret[0] = '"'
	copy(ret[1:], []byte(str))
	ret[len(ret)-1] = '"'
	return ret, nil
}

// UnmarshalJSON convert t from its JSON representation
func (t *ProtoBigInt) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return ErrInvalidValue
	}
	if _, ok := t.i.SetString(string(data[1:len(data)-1]), 10); !ok {
		return ErrInvalidValue
	}
	return nil

}

// Compare only required if the compare option is set
func (t ProtoBigInt) Compare(other ProtoBigInt) int {
	return t.i.Cmp(&other.i)
}

// Equal only required if the equal option is set
func (t ProtoBigInt) Equal(other ProtoBigInt) bool {
	return t.i.Cmp(&other.i) == 0
}

// Get returns a pointer to the inner math/big.Int
func (t *ProtoBigInt) Get() *big.Int {
	return &t.i
}

// Set sets the value of the inner big.Int to p
func (t *ProtoBigInt) Set(p *big.Int) {
	t.i.Set(p)
}
