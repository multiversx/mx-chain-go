package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
)

type HeartbeatStorerStub struct {
	LoadGenesisTimeCalled   func() (time.Time, error)
	UpdateGenesisTimeCalled func(genesisTime time.Time) error
	LoadHbmiDTOCalled       func(pubKey string) (*heartbeat.HeartbeatDTO, error)
	SavePubkeyDataCalled    func(pubkey []byte, heartbeat *heartbeat.HeartbeatDTO) error
	LoadKeysCalled          func() ([][]byte, error)
	SaveKeysCalled          func(peersSlice [][]byte) error
}

func (hss *HeartbeatStorerStub) LoadGenesisTime() (time.Time, error) {
	return hss.LoadGenesisTimeCalled()
}

func (hss *HeartbeatStorerStub) UpdateGenesisTime(genesisTime time.Time) error {
	return hss.UpdateGenesisTimeCalled(genesisTime)
}

func (hss *HeartbeatStorerStub) LoadHbmiDTO(pubKey string) (*heartbeat.HeartbeatDTO, error) {
	return hss.LoadHbmiDTOCalled(pubKey)
}

func (hss *HeartbeatStorerStub) SavePubkeyData(pubkey []byte, heartbeat *heartbeat.HeartbeatDTO) error {
	return hss.SavePubkeyDataCalled(pubkey, heartbeat)
}

func (hss *HeartbeatStorerStub) LoadKeys() ([][]byte, error) {
	return hss.LoadKeysCalled()
}

func (hss *HeartbeatStorerStub) SaveKeys(peersSlice [][]byte) error {
	return hss.SaveKeysCalled(peersSlice)
}

func (hss *HeartbeatStorerStub) IsInterfaceNil() bool {
	return false
}
