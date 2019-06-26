package consensus

import (
	"math/big"
)

type validator struct {
	stake  *big.Int
	rating int32
	pubKey []byte
}

// NewValidator creates a new instance of a validator
func NewValidator(stake *big.Int, rating int32, pubKey []byte) (*validator, error) {
	if stake == nil {
		return nil, ErrNilStake
	}

	if stake.Cmp(big.NewInt(0)) < 0 {
		return nil, ErrNegativeStake
	}

	if pubKey == nil {
		return nil, ErrNilPubKey
	}

	return &validator{
		stake:  stake,
		rating: rating,
		pubKey: pubKey,
	}, nil
}

// Stake returns the validator's stake
func (v *validator) Stake() *big.Int {
	return v.stake
}

// Rating returns the validator's rating
func (v *validator) Rating() int32 {
	return v.rating
}

// PubKey returns the validator's public key
func (v *validator) PubKey() []byte {
	return v.pubKey
}
