package crypto

// Key represents a crypto key - can be of the private or public type
type Key interface {
	// ToByteArray returns the byte array representation of the key
	ToByteArray() ([]byte, error)
}

// PrivateKey represents a private key that can sign data or decrypt messages encrypted with a public key
type PrivateKey interface {
	Key
	// Sign can be used to sign a message with the private key
	Sign(message []byte) error
}

// PublicKey can be used to encrypt messages
type PublicKey interface {
	Key
	// Verify signature represents the signed hash of the data
	Verify(data []byte, signature []byte) (bool, error)
}



