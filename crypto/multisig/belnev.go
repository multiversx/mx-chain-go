package multisig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

type belNev struct {
	message           []byte
	pubKeys           []crypto.PublicKey
	privKey           crypto.PrivateKey
	commHashes        [][]byte
	commHashesBitmap  []byte
	commSecret        []byte
	commitments       [][]byte
	commitmentsBitmap []byte
	aggCommitment     []byte
	sigShares         [][]byte
	sigSharesBitmap   []byte
	aggSig            []byte
	ownIndex          uint16
	hasher            hashing.Hasher
}

// NewBelNevMultisig creates a new Belare Neven multi-signer
func NewBelNevMultisig(
	hasher hashing.Hasher,
	pubKeys []string,
	privKey crypto.PrivateKey,
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

	var pk []crypto.PublicKey
	kg := schnorr.NewKeyGenerator()

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

	// own index is used only for signing
	return &belNev{
		pubKeys:  pk,
		privKey:  privKey,
		ownIndex: ownIndex,
		hasher:   hasher,
	}, nil
}

// NewMultiSiger instantiates another multiSigner of the same type
func (bn *belNev) NewMultiSiger(hasher hashing.Hasher, pubKeys []string, key crypto.PrivateKey, index uint16) (crypto.MultiSigner, error) {
	return NewBelNevMultisig(hasher, pubKeys, key, index)
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
		return
	}

	commitment, err = pk.ToByteArray()

	return
}

// SetCommitmentSecret sets the committment secret
func (bn *belNev) SetCommitmentSecret(commSecret []byte) error {
	// TODO

	return nil
}

// CommitmentBitmap returns the bitmap with the set
func (bn *belNev) CommitmentBitmap() []byte {
	// TODO

	return []byte("implement me")
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
func (bn *belNev) AggregateCommitments() ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// SetAggCommitment sets the aggregated commitment for the marked signers in bitmap
func (bn *belNev) SetAggCommitment(aggCommitment []byte, bitmap []byte) error {
	// TODO

	return nil
}

// SignPartial creates a partial signature
func (bn *belNev) SignPartial() ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// SigBitmap returns the bitmap for the set partial signatures
func (bn *belNev) SigBitmap() []byte {
	// TODO

	return []byte("implement me")
}

// SetSigBitmap sets the signers bitmap. Starting with index 0, each signer has 1 bit according to it's position in the
// signers list, set to 1 if signer's signature is used, 0 if not used
func (bn *belNev) SetSigBitmap([]byte) error {
	// TODO

	return nil
}

// VerifyPartial verifies the partial signature of the signer with specified position
func (bn *belNev) VerifyPartial(index uint16, sig []byte) error {
	// TODO

	return nil
}

// AddSignPartial adds the partial signature of the signer with specified position
func (bn *belNev) AddSignPartial(index uint16, sig []byte) error {
	// TODO

	return nil
}

// AggregateSigs aggregates all collected partial signatures
func (bn *belNev) AggregateSigs() ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// SetAggregatedSig sets the aggregated signature
func (bn *belNev) SetAggregatedSig([]byte) error {
	// TODO

	return nil
}

// Verify verifies the aggregated signature
func (bn *belNev) Verify() error {
	// TODO

	return nil
}
