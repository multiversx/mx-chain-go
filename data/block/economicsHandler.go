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

	e.TotalSupply = totalSupply
	return nil
}

// SetTotalToDistribute sets the total to be distributed
func (e *Economics) SetTotalToDistribute(totalToDistribute *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.TotalToDistribute = totalToDistribute
	return nil
}

// SetTotalNewlyMinted sets the total number of minted tokens
func (e *Economics) SetTotalNewlyMinted(totalNewlyMinted *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.TotalNewlyMinted = totalNewlyMinted
	return nil
}

// SetRewardsPerBlock sets the rewards per block
func (e *Economics) SetRewardsPerBlock(rewardsPerBlock *big.Int) error {
	if e == nil {
		 return data.ErrNilPointerReceiver
	}

	e.RewardsPerBlock = rewardsPerBlock
	return nil
}

// SetRewardsForProtocolSustainability sets the rewards for protocol sustainability
func (e *Economics) SetRewardsForProtocolSustainability(rewardsForProtocolSustainability *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.RewardsForProtocolSustainability = rewardsForProtocolSustainability
	return nil
}

// SetNodePrice sets the node price
func (e *Economics) SetNodePrice(nodePrice *big.Int) error {
	if e == nil {
		return data.ErrNilPointerReceiver
	}

	e.NodePrice = nodePrice
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
