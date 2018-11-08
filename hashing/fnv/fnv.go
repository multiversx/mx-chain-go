package fnv

import (
	"hash/fnv"
)

var fnvEmptyHash []byte

type Fnv struct {
}

func (Fnv) Compute(s string) []byte {
	h := fnv.New128a()
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (f Fnv) EmptyHash() []byte {
	if len(fnvEmptyHash) == 0 {
		fnvEmptyHash = f.Compute("")
	}
	return fnvEmptyHash
}

func (Fnv) Size() int {
	return fnv.New128a().Size()
}
