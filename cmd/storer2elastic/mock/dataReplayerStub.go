package mock

import (
	storer2ElasticData "github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
)

// DataReplayerStub -
type DataReplayerStub struct {
	RangeCalled func(handler func(persistedData storer2ElasticData.RoundPersistedData) bool) error
}

// Range -
func (d *DataReplayerStub) Range(handler func(persistedData storer2ElasticData.RoundPersistedData) bool) error {
	if d.RangeCalled != nil {
		return d.RangeCalled(handler)
	}

	return nil
}

// IsInterfaceNil -
func (d *DataReplayerStub) IsInterfaceNil() bool {
	return d == nil
}
