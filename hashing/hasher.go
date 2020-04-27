package hashing

// BlsHashSize specifies the hash size for using bls scheme
const BlsHashSize = 16

// Hasher provides hashing services
type Hasher interface {
	Compute(string) []byte
	EmptyHash() []byte
	Size() int
	IsInterfaceNil() bool
}
