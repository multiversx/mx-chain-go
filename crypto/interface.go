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
	Verify(data []byte, signature []byte) error
}

// MultiSigner provides functionality for multi-signing a message
type MultiSigner interface {
	// MultiSigVerifier Provides functionality for verifying a multi-signature
	MultiSigVerifier
	// CreateCommitment creates a secret commitment and the corresponding public commitment point
	CreateCommitment() (commSecret []byte, commitment []byte, err error)
	// AddCommitmentHash adds a commitment hash to the list with the specified position
	AddCommitmentHash(index uint16, commHash []byte) error
	// CommitmentHash returns the commitment hash from the list with the specified position
	CommitmentHash(index uint16) ([]byte, error)
	// SetCommitmentSecret sets the committment secret
	SetCommitmentSecret(commSecret []byte) error
	// AddCommitment adds a commitment to the list with the specified position
	AddCommitment(index uint16, value []byte) error
	// Commitment returns the commitment from the list with the specified position
	Commitment(index uint16) ([]byte, error)
	// AggregateCommitments aggregates the list of commitments
	AggregateCommitments(bitmap []byte) ([]byte, error)
	// SetAggCommitment sets the aggregated commitment
	SetAggCommitment(aggCommitment []byte) error
	// CreateSignatureShare creates a partial signature
	CreateSignatureShare(bitmap []byte) ([]byte, error)
	// AddSignatureShare adds the partial signature of the signer with specified position
	AddSignatureShare(index uint16, sig []byte) error
	// SignatureShare returns the partial signature set for given index
	SignatureShare(index uint16) ([]byte, error)
	// VerifySignatureShare verifies the partial signature of the signer with specified position
	VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error
	// AggregateSigs aggregates all collected partial signatures
	AggregateSigs(bitmap []byte) ([]byte, error)
}

// MultiSigVerifier Provides functionality for verifying a multi-signature
type MultiSigVerifier interface {
	// Reset resets the multisigner and initializes to the new params
	Reset(pubKeys []string, index uint16) error
	// SetMessage sets the message to be multi-signed upon
	SetMessage(msg []byte)
	// SetAggregatedSig sets the aggregated signature
	SetAggregatedSig([]byte) error
	// Verify verifies the aggregated signature
	Verify(bitmap []byte) error
}
