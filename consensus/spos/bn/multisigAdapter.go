package bn

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
)

type bnMultiSigner interface {
	crypto.MultiSigner
	// CreateCommitment creates a secret commitment and the corresponding public commitment point
	CreateCommitment() (commSecret []byte, commitment []byte)
	// StoreCommitmentHash adds a commitment hash to the list with the specified position
	StoreCommitmentHash(index uint16, commHash []byte) error
	// CommitmentHash returns the commitment hash from the list with the specified position
	CommitmentHash(index uint16) ([]byte, error)
	// StoreCommitment adds a commitment to the list with the specified position
	StoreCommitment(index uint16, value []byte) error
	// Commitment returns the commitment from the list with the specified position
	Commitment(index uint16) ([]byte, error)
	// AggregateCommitments aggregates the list of commitments
	AggregateCommitments(bitmap []byte) error
}

// getBnMultiSigner returns the belare neven multi-signer
func getBnMultiSigner(musig crypto.MultiSigner) (bnMultiSigner, error) {
	bnMultiSig, ok := musig.(bnMultiSigner)
	if !ok {
		return nil, spos.ErrInvalidMultiSigner
	}

	return bnMultiSig, nil
}
