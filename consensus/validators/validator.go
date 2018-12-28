package validators

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
)

type validator struct {
	stake  big.Int
	rating int32
	pubKey []byte
}

func NewValidator(stake big.Int, rating int32, pubKey []byte) (*validator, error) {
	if stake.Cmp(big.NewInt(0)) < 0 {
		return nil, consensus.ErrNegativeStake
	}

	if pubKey == nil {
		return nil, consensus.ErrNilPubKey
	}

	return &validator{
		stake:  stake,
		rating: rating,
		pubKey: pubKey,
	}, nil
}

func (v *validator) Stake() big.Int {
	return v.stake
}

func (v *validator) Rating() int32 {
	return v.rating
}

func (v *validator) PubKey() []byte {
	return v.pubKey
}
