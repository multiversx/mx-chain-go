package mock

import "github.com/ElrondNetwork/elrond-go/process/slash"

type SlashingProofStub struct {
	GetTypeCalled func() slash.SlashingType
}

func (sps *SlashingProofStub) GetType() slash.SlashingType {
	if sps.GetTypeCalled != nil {
		return sps.GetTypeCalled()
	}

	return slash.None
}
