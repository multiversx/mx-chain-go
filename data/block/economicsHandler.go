package block

import "math/big"

// SetTotalSupply -
func (m *Economics) SetTotalSupply(totalSupply *big.Int) {
	if m != nil {
		return
	}
	m.TotalSupply = totalSupply
}

// SetTotalToDistribute -
func (m *Economics) SetTotalToDistribute(totalToDistribute *big.Int) {
	if m != nil {
		return
	}
	m.TotalToDistribute = totalToDistribute
}

// SetTotalNewlyMinted -
func (m *Economics) SetTotalNewlyMinted(totalNewlyMinted *big.Int) {
	if m != nil {
		return
	}
	m.TotalNewlyMinted = totalNewlyMinted
}

// SetRewardsPerBlock -
func (m *Economics) SetRewardsPerBlock(rewardsPerBlock *big.Int) {
	if m != nil {
		return
	}
	m.RewardsPerBlock = rewardsPerBlock
}

// SetRewardsForProtocolSustainability -
func (m *Economics) SetRewardsForProtocolSustainability(rewardsForProtocolSustainability *big.Int) {
	if m != nil {
		return
	}
	m.RewardsForProtocolSustainability = rewardsForProtocolSustainability
}

// SetNodePrice -
func (m *Economics) SetNodePrice(nodePrice *big.Int) {
	if m != nil {
		return
	}
	m.NodePrice = nodePrice
}

// SetPrevEpochStartRound -
func (m *Economics) SetPrevEpochStartRound(prevEpochStartRound uint64) {
	if m != nil {
		return
	}
	m.PrevEpochStartRound = prevEpochStartRound
}

// SetPrevEpochStartHash -
func (m *Economics) SetPrevEpochStartHash(prevEpochStartHash []byte) {
	if m != nil {
		return
	}
	m.PrevEpochStartHash = prevEpochStartHash
}
