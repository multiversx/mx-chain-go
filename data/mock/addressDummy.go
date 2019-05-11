package mock

type AddressDummy struct {
	bytes []byte
	hash  []byte
}

func NewAddressDummy(bytes, hash []byte) *AddressDummy {
	return &AddressDummy{
		bytes: bytes,
		hash:  hash,
	}
}

func (ad *AddressDummy) Bytes() []byte {
	return ad.bytes
}

func (ad *AddressDummy) Hash() []byte {
	return ad.hash
}
