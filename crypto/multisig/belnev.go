package multisig

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
)

type belNev struct {
	message           []byte
	pubKeys           []crypto.PublicKey
	privKey           crypto.PrivateKey
	commSecret        []byte
	commitments       [][]byte
	commitmentsBitmap []byte
	aggCommitment     []byte
	challenges        [][]byte
	sigShares         []byte
	sigSharesBitmap   []byte
	aggSig            []byte
	ownIndex          uint16
}

// NewBelNevMultisig creates a new Belare Neven multi-signer
func NewBelNevMultisig(pubKeys []crypto.PublicKey, privKey crypto.PrivateKey, ownIndex uint16) (*belNev, error) {
	if privKey == nil {
		return nil, crypto.ErrNilPrivateKey
	}

	if pubKeys == nil {
		return nil, crypto.ErrNilPublicKeys
	}

	// own index is used only for signing
	return &belNev{
		pubKeys:  pubKeys,
		privKey:  privKey,
		ownIndex: ownIndex,
	}, nil
}

// SetMessage sets the message to be multi-signed upon
func (bn *belNev) SetMessage(msg []byte) {
	bn.message = msg
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

// SetSigBitmap sets the bitmap for the participating signers
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
