package mock

// ChanceComputerStub -
type ChanceComputerStub struct {
	GetChanceCalled func(rating uint32) uint32
}

// GetChance -
func (c *ChanceComputerStub) GetChance(rating uint32) uint32 {
	if c.GetChanceCalled != nil {
		return c.GetChanceCalled(rating)
	}

	return 80
}

// IsInterfaceNil -
func (c *ChanceComputerStub) IsInterfaceNil() bool {
	return c == nil
}
