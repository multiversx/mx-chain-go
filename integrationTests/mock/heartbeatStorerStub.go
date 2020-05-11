package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
)

// HeartbeatStorerStub -
type HeartbeatStorerStub struct {
	LoadGenesisTimeCalled   func() (time.Time, error)
	UpdateGenesisTimeCalled func(genesisTime time.Time) error
	LoadHeartBeatDTOCalled  func(pubKey string) (*data.HeartbeatDTO, error)
	SavePubkeyDataCalled    func(pubkey []byte, heartbeat *data.HeartbeatDTO) error
	LoadKeysCalled          func() ([][]byte, error)
	SaveKeysCalled          func(peersSlice [][]byte) error
}

// LoadGenesisTime -
func (hss *HeartbeatStorerStub) LoadGenesisTime() (time.Time, error) {
	return hss.LoadGenesisTimeCalled()
}

// UpdateGenesisTime -
func (hss *HeartbeatStorerStub) UpdateGenesisTime(genesisTime time.Time) error {
	return hss.UpdateGenesisTimeCalled(genesisTime)
}

// LoadHeartBeatDTO -
func (hss *HeartbeatStorerStub) LoadHeartBeatDTO(pubKey string) (*data.HeartbeatDTO, error) {
	return hss.LoadHeartBeatDTOCalled(pubKey)
}

// SavePubkeyData -
func (hss *HeartbeatStorerStub) SavePubkeyData(pubkey []byte, heartbeat *data.HeartbeatDTO) error {
	return hss.SavePubkeyDataCalled(pubkey, heartbeat)
}

// LoadKeys -
func (hss *HeartbeatStorerStub) LoadKeys() ([][]byte, error) {
	return hss.LoadKeysCalled()
}

// SaveKeys -
func (hss *HeartbeatStorerStub) SaveKeys(peersSlice [][]byte) error {
	return hss.SaveKeysCalled(peersSlice)
}

// IsInterfaceNil -
func (hss *HeartbeatStorerStub) IsInterfaceNil() bool {
	return false
}
