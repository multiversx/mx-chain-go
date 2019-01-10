package multisig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

type belNev struct {
	message       []byte
	pubKeys       []crypto.PublicKey
	privKey       crypto.PrivateKey
	commHashes    [][]byte
	commSecret    []byte
	commitments   [][]byte
	aggCommitment []byte
	sigShares     [][]byte
	aggSig        []byte
	ownIndex      uint16
	hasher        hashing.Hasher
	keyGen        crypto.KeyGenerator
}

// NewBelNevMultisig creates a new Bellare Neven multi-signer
func NewBelNevMultisig(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
	keyGen crypto.KeyGenerator,
	ownIndex uint16) (*belNev, error) {

	if hasher == nil {
		return nil, crypto.ErrNilHasher
	}

	if privKey == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	if pubKeys == nil {
		return nil, crypto.ErrNilPublicKeys
	}

	if keyGen == nil {
		return nil, crypto.ErrNilKeyGenerator
	}

	pk, err := convertStringsToPubKeys(pubKeys, keyGen)

	if err != nil {
		return nil, err
	}

	// own index is used only for signing
	return &belNev{
		pubKeys:  pk,
		privKey:  privKey,
		ownIndex: ownIndex,
		hasher:   hasher,
		keyGen:   keyGen,
	}, nil
}

func convertStringsToPubKeys(pubKeys []string, kg crypto.KeyGenerator) ([]crypto.PublicKey, error) {
	var pk []crypto.PublicKey

	//convert pubKeys
	for _, pubKeyStr := range pubKeys {
		if pubKeyStr != "" {
			pubKey, err := kg.PublicKeyFromByteArray([]byte(pubKeyStr))

			if err != nil {
				return nil, crypto.ErrInvalidPublicKeyString
			}

			pk = append(pk, pubKey)
		} else {
			return nil, crypto.ErrNilPublicKey
		}
	}
	return pk, nil
}

// Reset resets the multiSigner and initializes corresponding fields with the given params
func (bn *belNev) Reset(pubKeys []string, index uint16) error {
	pk, err := convertStringsToPubKeys(pubKeys, bn.keyGen)

	if err != nil {
		return err
	}

	bn.message = nil
	bn.ownIndex = index
	bn.pubKeys = pk
	bn.commSecret = nil
	bn.commitments = nil
	bn.aggCommitment = nil
	bn.aggSig = nil
	bn.commHashes = nil
	bn.sigShares = nil

	return nil
}

// SetMessage sets the message to be multi-signed upon
func (bn *belNev) SetMessage(msg []byte) {
	bn.message = msg
}

// AddCommitmentHash sets a commitment Hash
func (bn *belNev) AddCommitmentHash(index uint16, commHash []byte) error {
	// TODO

	return nil
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bn *belNev) CommitmentHash(index uint16) ([]byte, error) {
	if int(index) >= len(bn.commHashes) {
		return nil, crypto.ErrInvalidIndex
	}

	if bn.commHashes[index] == nil {
		return nil, crypto.ErrNilElement
	}

	return bn.commHashes[index], nil
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bn *belNev) CreateCommitment() (commSecret []byte, commitment []byte, err error) {
	kg := schnorr.NewKeyGenerator()
	sk, pk := kg.GeneratePair()

	commSecret, err = sk.ToByteArray()

	if err != nil {
		return nil, nil, err
	}

	commitment, err = pk.ToByteArray()

	if err != nil {
		return nil, nil, err
	}

	return commSecret, commitment, nil
}

// SetCommitmentSecret sets the committment secret
func (bn *belNev) SetCommitmentSecret(commSecret []byte) error {
	// TODO

	return nil
}

// AddCommitment adds a commitment to the list on the specified position
func (bn *belNev) AddCommitment(index uint16, value []byte) error {
	// TODO

	return nil
}

// Commitment returns the commitment from the list with the specified position
func (bn *belNev) Commitment(index uint16) ([]byte, error) {
	if int(index) >= len(bn.commitments) {
		return nil, crypto.ErrInvalidIndex
	}

	if bn.commitments[index] == nil {
		return nil, crypto.ErrNilElement
	}

	return bn.commitments[index], nil
}

// AggregateCommitments aggregates the list of commitments
func (bn *belNev) AggregateCommitments(bitmap []byte) ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// SetAggCommitment sets the aggregated commitment for the marked signers in bitmap
func (bn *belNev) SetAggCommitment(aggCommitment []byte) error {
	// TODO

	return nil
}

// SignPartial creates a partial signature
func (bn *belNev) SignPartial(bitmap []byte) ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// VerifyPartial verifies the partial signature of the signer with specified position
func (bn *belNev) VerifyPartial(index uint16, sig []byte, bitmap []byte) error {
	// TODO

	return nil
}

// AddSignPartial adds the partial signature of the signer with specified position
func (bn *belNev) AddSignPartial(index uint16, sig []byte) error {
	// TODO

	return nil
}

// AggregateSigs aggregates all collected partial signatures
func (bn *belNev) AggregateSigs(bitmap []byte) ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// SetAggregatedSig sets the aggregated signature
func (bn *belNev) SetAggregatedSig([]byte) error {
	// TODO

	return nil
}

// Verify verifies the aggregated signature
func (bn *belNev) Verify(bitmap []byte) error {
	// TODO

	return nil
}
