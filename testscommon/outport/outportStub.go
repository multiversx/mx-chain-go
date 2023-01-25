package outport

import (
	"github.com/multiversx/mx-chain-core-go/data"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/outport"
)

// OutportStub is a mock implementation fot the OutportHandler interface
type OutportStub struct {
	SaveBlockCalled             func(args *outportcore.ArgsSaveBlockData)
	SaveValidatorsRatingCalled  func(index string, validatorsInfo []*outportcore.ValidatorRatingInfo)
	SaveValidatorsPubKeysCalled func(shardPubKeys map[uint32][][]byte, epoch uint32)
	HasDriversCalled            func() bool
}

// SaveBlock -
func (as *OutportStub) SaveBlock(args *outportcore.ArgsSaveBlockData) {
	if as.SaveBlockCalled != nil {
		as.SaveBlockCalled(args)
	}
}

// SaveValidatorsRating -
func (as *OutportStub) SaveValidatorsRating(index string, validatorsInfo []*outportcore.ValidatorRatingInfo) {
	if as.SaveValidatorsRatingCalled != nil {
		as.SaveValidatorsRatingCalled(index, validatorsInfo)
	}
}

// SaveValidatorsPubKeys -
func (as *OutportStub) SaveValidatorsPubKeys(shardPubKeys map[uint32][][]byte, epoch uint32) {
	if as.SaveValidatorsPubKeysCalled != nil {
		as.SaveValidatorsPubKeysCalled(shardPubKeys, epoch)
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (as *OutportStub) IsInterfaceNil() bool {
	return as == nil
}

// HasDrivers -
func (as *OutportStub) HasDrivers() bool {
	if as.HasDriversCalled != nil {
		return as.HasDriversCalled()
	}
	return false
}

// RevertIndexedBlock -
func (as *OutportStub) RevertIndexedBlock(_ data.HeaderHandler, _ data.BodyHandler) {

}

// SaveAccounts -
func (as *OutportStub) SaveAccounts(_ uint64, _ map[string]*outportcore.AlteredAccount, _ uint32) {

}

// Close -
func (as *OutportStub) Close() error {
	return nil
}

// SaveRoundsInfo -
func (as *OutportStub) SaveRoundsInfo(_ []*outportcore.RoundInfo) {

}

// SubscribeDriver -
func (as *OutportStub) SubscribeDriver(_ outport.Driver) error {
	return nil
}

// FinalizedBlock -
func (as *OutportStub) FinalizedBlock(_ []byte) {
}
