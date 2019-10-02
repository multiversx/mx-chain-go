package storage

import (
	"errors"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"time"

	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

const peersKeysDbEntry = "keys"
const genesisTimeDbEntry = "genesisTime"

type HeartbeatDbStorer struct {
	storer      storage.Storer
	marshalizer marshal.Marshalizer
}

// TODO: add comments and tests

func NewHeartbeatStorer(
	storer storage.Storer,
	marshalizer marshal.Marshalizer,
) (*HeartbeatDbStorer, error) {
	if storer == nil || storer.IsInterfaceNil() {
		return nil, heartbeat.ErrNilMonitorDb
	}
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, heartbeat.ErrNilMarshalizer
	}

	return &HeartbeatDbStorer{
		storer:      storer,
		marshalizer: marshalizer,
	}, nil
}

func (hs *HeartbeatDbStorer) LoadGenesisTime() (time.Time, error) {
	return time.Now(), nil
}

func (hs *HeartbeatDbStorer) UpdateGenesisTime(genesisTime time.Time) error {
	if hs.storer.Has([]byte(genesisTimeDbEntry)) != nil {
		err := hs.saveGenesisTimeToDb(genesisTime)
		if err != nil {
			return err
		}
	} else {
		genesisTimeFromDbBytes, err := hs.storer.Get([]byte(genesisTimeDbEntry))
		if err != nil {
			return errors.New("monitor: can't get genesis time from db")
		}

		var genesisTimeFromDb time.Time
		err = hs.marshalizer.Unmarshal(&genesisTimeFromDb, genesisTimeFromDbBytes)
		if err != nil {
			return errors.New("monitor: can't unmarshal genesis time")
		}

		if genesisTimeFromDb != genesisTime {
			log.Info(fmt.Sprintf("updated heartbeat's genesis time to %s", genesisTimeFromDb))
		}

		err = hs.saveGenesisTimeToDb(genesisTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func (hs *HeartbeatDbStorer) saveGenesisTimeToDb(genesisTime time.Time) error {
	genesisTimeBytes, err := hs.marshalizer.Marshal(genesisTime)
	if err != nil {
		return errors.New("monitor: can't marshal genesis time")
	}

	err = hs.storer.Put([]byte(genesisTimeDbEntry), genesisTimeBytes)
	if err != nil {
		return errors.New("monitor: can't store genesis time")
	}

	return nil
}

func (hs *HeartbeatDbStorer) LoadPubkeysData() (map[string]*heartbeat.HeartbeatDTO, error) {
	return nil, nil
}

func (hs *HeartbeatDbStorer) LoadHbmiDTO(pubKey string) (*heartbeat.HeartbeatDTO, error) {
	pkbytes := []byte(pubKey)

	hbFromDB, err := hs.storer.Get(pkbytes)
	if err != nil {
		return nil, err
	}

	heartbeatDto := heartbeat.HeartbeatDTO{}
	err = hs.marshalizer.Unmarshal(&heartbeatDto, hbFromDB)
	if err != nil {
		return nil, err
	}

	return &heartbeatDto, nil
}

func (hs *HeartbeatDbStorer) LoadKeys() ([][]byte, error) {
	allKeysBytes, err := hs.storer.Get([]byte(peersKeysDbEntry))
	if err != nil {
		return nil, err
	}

	var peersSlice [][]byte
	err = hs.marshalizer.Unmarshal(&peersSlice, allKeysBytes)
	if err != nil {
		return nil, err
	}

	return peersSlice, nil
}

func (hs *HeartbeatDbStorer) SaveKeys(peersSlice [][]byte) error {
	marshalizedFullPeersSlice, errMarsh := hs.marshalizer.Marshal(peersSlice)
	if errMarsh != nil {
		return errMarsh
	}

	return hs.storer.Put([]byte(peersKeysDbEntry), marshalizedFullPeersSlice)
}

func (hs *HeartbeatDbStorer) SavePubkeyData(
	pubkey []byte,
	heartbeat *heartbeat.HeartbeatDTO,
) error {
	marshalizedHeartBeat, err := hs.marshalizer.Marshal(heartbeat)
	if err != nil {
		return err
	}

	errStore := hs.storer.Put(pubkey, marshalizedHeartBeat)
	if errStore != nil {
		return err
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (hs *HeartbeatDbStorer) IsInterfaceNil() bool {
	if hs == nil {
		return true
	}
	return false
}
