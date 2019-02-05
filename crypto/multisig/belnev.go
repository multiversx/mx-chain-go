package multisig

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/schnorr"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
)

type belNev struct {
	message        []byte
	pubKeys        []crypto.PublicKey
	privKey        crypto.PrivateKey
	mutCommHashes  *sync.RWMutex
	commHashes     [][]byte
	commSecret     []byte
	mutCommitments *sync.RWMutex
	commitments    [][]byte
	aggCommitment  []byte
	mutSigShares   *sync.RWMutex
	sigShares      [][]byte
	aggSig         []byte
	ownIndex       uint16
	hasher         hashing.Hasher
	keyGen         crypto.KeyGenerator
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

	sizeConsensus := len(pubKeys)

	commHashes := make([][]byte, sizeConsensus)
	commitments := make([][]byte, sizeConsensus)
	sigShares := make([][]byte, sizeConsensus)

	pk, err := convertStringsToPubKeys(pubKeys, keyGen)

	if err != nil {
		return nil, err
	}

	// own index is used only for signing
	return &belNev{
		pubKeys:        pk,
		privKey:        privKey,
		ownIndex:       ownIndex,
		hasher:         hasher,
		keyGen:         keyGen,
		mutCommHashes:  &sync.RWMutex{},
		commHashes:     commHashes,
		mutCommitments: &sync.RWMutex{},
		commitments:    commitments,
		mutSigShares:   &sync.RWMutex{},
		sigShares:      sigShares,
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

	sizeConsensus := len(pubKeys)

	bn.message = nil
	bn.ownIndex = index
	bn.pubKeys = pk
	bn.commSecret = nil
	bn.aggCommitment = nil
	bn.aggSig = nil

	bn.mutCommHashes.Lock()
	bn.commHashes = make([][]byte, sizeConsensus)
	bn.mutCommHashes.Unlock()

	bn.mutCommitments.Lock()
	bn.commitments = make([][]byte, sizeConsensus)
	bn.mutCommitments.Unlock()

	bn.mutSigShares.Lock()
	bn.sigShares = make([][]byte, sizeConsensus)
	bn.mutSigShares.Unlock()

	return nil
}

// SetMessage sets the message to be multi-signed upon
func (bn *belNev) SetMessage(msg []byte) {
	bn.message = msg
}

// AddCommitmentHash sets a commitment Hash
func (bn *belNev) AddCommitmentHash(index uint16, commHash []byte) error {
	bn.mutCommHashes.Lock()
	if int(index) >= len(bn.commHashes) {
		bn.mutCommHashes.Unlock()
		return crypto.ErrInvalidIndex
	}

	bn.commHashes[index] = commHash
	bn.mutCommHashes.Unlock()
	return nil
}

// CommitmentHash returns the commitment hash from the list on the specified position
func (bn *belNev) CommitmentHash(index uint16) ([]byte, error) {
	bn.mutCommHashes.RLock()
	defer bn.mutCommHashes.RUnlock()

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
	bn.commSecret = commSecret
	return nil
}

// AddCommitment adds a commitment to the list on the specified position
func (bn *belNev) AddCommitment(index uint16, commitment []byte) error {
	bn.mutCommitments.Lock()
	if int(index) >= len(bn.commitments) {
		bn.mutCommitments.Unlock()
		return crypto.ErrInvalidIndex
	}

	bn.commitments[index] = commitment
	bn.mutCommitments.Unlock()
	return nil
}

// Commitment returns the commitment from the list with the specified position
func (bn *belNev) Commitment(index uint16) ([]byte, error) {
	bn.mutCommitments.RLock()
	defer bn.mutCommitments.RUnlock()

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
	// TODO, do not forget about mutCommitments

	return []byte("implement me"), nil
}

// SetAggCommitment sets the aggregated commitment for the marked signers in bitmap
func (bn *belNev) SetAggCommitment(aggCommitment []byte) error {
	bn.aggCommitment = aggCommitment
	return nil
}

// CreateSignatureShare creates a partial signature
func (bn *belNev) CreateSignatureShare(bitmap []byte) ([]byte, error) {
	// TODO

	return []byte("implement me"), nil
}

// VerifySignatureShare verifies the partial signature of the signer with specified position
func (bn *belNev) VerifySignatureShare(index uint16, sig []byte, bitmap []byte) error {
	// TODO

	return nil
}

// AddSignatureShare adds the partial signature of the signer with specified position
func (bn *belNev) AddSignatureShare(index uint16, sig []byte) error {
	bn.mutSigShares.Lock()
	if int(index) >= len(bn.sigShares) {
		bn.mutSigShares.Unlock()
		return crypto.ErrInvalidIndex
	}

	bn.sigShares[index] = sig
	bn.mutSigShares.Unlock()
	return nil
}

// SignatureShare returns the partial signature set for given index
func (bn *belNev) SignatureShare(index uint16) ([]byte, error) {
	bn.mutSigShares.RLock()
	defer bn.mutSigShares.RUnlock()

	if int(index) >= len(bn.sigShares) {
		return nil, crypto.ErrInvalidIndex
	}

	if bn.sigShares[index] == nil {
		return nil, crypto.ErrNilElement
	}

	return bn.sigShares[index], nil
}

// AggregateSigs aggregates all collected partial signatures
func (bn *belNev) AggregateSigs(bitmap []byte) ([]byte, error) {
	// TODO, do not forget about mutSigShares

	return []byte("implement me"), nil
}

// SetAggregatedSig sets the aggregated signature
func (bn *belNev) SetAggregatedSig(aggSig []byte) error {
	bn.aggSig = aggSig
	return nil
}

// Verify verifies the aggregated signature
func (bn *belNev) Verify(bitmap []byte) error {
	// TODO

	return nil
}
