package crypto

// KeyGenerator is an interface for generating different types of cryptographic keys
type KeyGenerator interface {
	GeneratePair() (PrivateKey, PublicKey)
	PrivateKeyFromByteArray(b []byte) (PrivateKey, error)
	PublicKeyFromByteArray(b []byte) (PublicKey, error)
}

// Key represents a crypto key - can be either private or public
type Key interface {
	// ToByteArray returns the byte array representation of the key
	ToByteArray() ([]byte, error)
}

// PrivateKey represents a private key that can sign data or decrypt messages encrypted with a public key
type PrivateKey interface {
	Key
	// Sign can be used to sign a message with the private key
	Sign(message []byte) ([]byte, error)
	// GeneratePublic builds a public key for the current private key
	GeneratePublic() PublicKey
}

// PublicKey can be used to encrypt messages
type PublicKey interface {
	Key
	// Verify signature represents the signed hash of the data
	Verify(data []byte, signature []byte) (bool, error)
}



