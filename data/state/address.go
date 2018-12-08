package state

// address is the struct holding an Elrond address identifier
type address struct {
	bytes []byte
	hash  []byte
}

// newAddress creates a new Address with the same byte slice as the parameter received
func newAddress(adr []byte, hash []byte) *address {
	return &address{bytes: adr, hash: hash}
}

// Bytes returns the data corresponding to this address
func (adr *address) Bytes() []byte {
	return adr.bytes
}

// Hash return the address corresponding hash
func (adr *address) Hash() []byte {
	return adr.hash
}
