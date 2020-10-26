package legacyKeccak256

import "golang.org/x/crypto/sha3"

// TODO: finish this implementation and add unit tests

type LegacyKeccak256 struct {
}

func (l LegacyKeccak256) Compute(s string) []byte {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(s))
	resultedHash := hasher.Sum(nil)

	return resultedHash
}

func (l LegacyKeccak256) EmptyHash() []byte {
	return []byte{}
}

func (l LegacyKeccak256) Size() int {
	return sha3.NewLegacyKeccak256().Size()
}

func (l LegacyKeccak256) IsInterfaceNil() bool {
	return false
}
