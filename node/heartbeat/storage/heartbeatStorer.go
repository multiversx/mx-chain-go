package storage

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type heartbeatStorer struct {
	storer      storage.Storer
	marshalizer marshal.Marshalizer
}

func NewHeartbeatStorer() (*heartbeatStorer, error) {

}

func (hs *heartbeatStorer) LoadGenesisTime() (time.Time, error) {

}

func (hs *heartbeatStorer) SaveGenesisTime(genesisTime time.Time) error {

}

func (hs *heartbeatStorer) LoadPubkeysData() (map[string]*heartbeat.HeartbeatDTO, error) {

}

func (hs *heartbeatStorer) SavePubkeyData(
	pubkey string,
	heartbeat *heartbeat.HeartbeatDTO,
) error {

}
