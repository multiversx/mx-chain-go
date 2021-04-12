package block

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data"
)

// SetTotalSupply sets the total supply
func (e *Economics) SetTotalSupply(totalSupply *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}
	if totalSupply == nil{
		return data.ErrInvalidValue
	}
	if e.TotalSupply == nil {
		e.TotalSupply = big.NewInt(0)
	}

	e.TotalSupply.Set(totalSupply)

	return nil
}

// SetTotalToDistribute sets the total to be distributed
func (e *Economics) SetTotalToDistribute(totalToDistribute *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}
	if totalToDistribute == nil {
		return data.ErrInvalidValue
	}
	if e.TotalToDistribute == nil {
		e.TotalToDistribute = big.NewInt(0)
	}

	e.TotalToDistribute.Set(totalToDistribute)

	return nil
}

// SetTotalNewlyMinted sets the total number of minted tokens
func (e *Economics) SetTotalNewlyMinted(totalNewlyMinted *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}
	if totalNewlyMinted == nil{
		return data.ErrInvalidValue
	}
	if e.TotalNewlyMinted == nil{
		e.TotalNewlyMinted = big.NewInt(0)
	}

	e.TotalNewlyMinted.Set(totalNewlyMinted)

	return nil
}

// SetRewardsPerBlock sets the rewards per block
func (e *Economics) SetRewardsPerBlock(rewardsPerBlock *big.Int) error {
	if e == nil {
		 return data.ErrNilPointerReceiver
	}
	if rewardsPerBlock == nil{
		return data.ErrInvalidValue
	}
	if e.RewardsPerBlock == nil{
		e.RewardsPerBlock = big.NewInt(0)
	}

	e.RewardsPerBlock.Set(rewardsPerBlock)

	return nil
}

// SetRewardsForProtocolSustainability sets the rewards for protocol sustainability
func (e *Economics) SetRewardsForProtocolSustainability(rewardsForProtocolSustainability *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}
	if rewardsForProtocolSustainability == nil{
		return data.ErrInvalidValue
	}
	if e.RewardsForProtocolSustainability == nil{
		e.RewardsForProtocolSustainability = big.NewInt(0)
	}

	e.RewardsForProtocolSustainability.Set(rewardsForProtocolSustainability)

	return nil
}

// SetNodePrice sets the node price
func (e *Economics) SetNodePrice(nodePrice *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}
	if nodePrice == nil{
		return data.ErrInvalidValue
	}
	if e.NodePrice == nil{
		e.NodePrice = big.NewInt(0)
	}

	e.NodePrice.Set(nodePrice)

	return nil
}

// SetPrevEpochStartRound sets the previous epoch start round
func (e *Economics) SetPrevEpochStartRound(prevEpochStartRound uint64) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.PrevEpochStartRound = prevEpochStartRound
	return nil
}

// SetPrevEpochStartHash sets the previous epoch start hash
func (e *Economics) SetPrevEpochStartHash(prevEpochStartHash []byte) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.PrevEpochStartHash = prevEpochStartHash
	return nil
}
