package multisig

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto/signing"
)

/*
Package belnev implements the multi-signature algorithm (BN-musig) presented in paper
"Multi-Signatures in the Plain Public-Key Model and a General Forking Lemma"
by Mihir Bellare and Gregory Neven. See https://cseweb.ucsd.edu/~mihir/papers/multisignatures-ccs.pdf.
This package provides the functionality for the cryptographic operations.
The message transfer functionality required for the algorithm are assumed to be
handled elsewhere. An overview of the protocol will be provided below.

The BN-musig protocol has 4 phases executed between a list of participants (public keys) L,
having a protocol leader (index = 0) and validators (index > 0). Each participant has it's
own private/public key pair (x_i, X_i), where x_i is the private key of participant
i and X_i is it's associated public key X_i = x_i * G. G is the base point on the used
curve, and * is the scalar multiplication. The protocol assumes that each participant has
the same view on the ordering of participants in the list L, so when communicating it is
not needed to pass on the list as well, but only a bitmap for the participating validators.

The phases of the protocol are as follows:

1. All participants/signers in L (including the leader) choose a random scalar 1 < r_i < n-1,
called a commitment secret, calculates the commitment R_i = r_i *G, and the commitment hash
t_i = H0(R_i), where H0 is a hashing function, different than H1 which will be used in next
rounds. Each of the members then broadcast these commitment hashes to all participants
including leader. (In a protocol where the communication is done only through the leader it needs
to be taken into account that the leader might not be honest)

2. When each signer i receives the commitment hash together with the public key of the sender
(t_j, X_j) it will send back the full R_i along with its public key, (R_i, X_i)

3. When signer i receives the full commitment from a signer j, it computes t_j = H0(R_j)
and verifies it with the previously received t_j. Locally each participant keeps track of
the sender, using a bitmap initially set to 0, by setting the corresponding bit to 1 and
storing the received commitment. If the commitment fails to validate the hash the protocol
is aborted. If commitment is not received in a bounded time delta and less than 2/3 of signers
have provided the commitment then protocol is aborted. If there are enough commitments >2/3
the leader broadcasts the bitmap.

4. When signer i receives the bitmap, it calculates the signature share and broadcasts it to all
participating signers. R = Sum(R_i) is the aggregated commitment each signer needs to calculate
before signing. s_i = r_i + H1(<L'>||X_i||R||m)*x_i is the signature share for signer i

When signer i receives all signature shares, it can calculate the aggregated signature
S=Sum(s_i) for all s_i of the participants.
*/

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
	kg := signing.NewKeyGenerator(bn.privKey.Suite())
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
