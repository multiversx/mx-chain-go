package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

const peersKeysDbEntry = "keys"
const genesisTimeDbEntry = "genesisTime"

// HeartbeatDbStorer is the struct which will handle storage operations for heartbeat
type HeartbeatDbStorer struct {
	storer      storage.Storer
	marshalizer marshal.Marshalizer
}

// NewHeartbeatDbStorer will create an instance of HeartbeatDbStorer
func NewHeartbeatDbStorer(
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

// LoadGenesisTime will return the genesis time saved in the storer
func (hs *HeartbeatDbStorer) LoadGenesisTime() (time.Time, error) {
	genesisTimeFromDbBytes, err := hs.storer.Get([]byte(genesisTimeDbEntry))
	if err != nil {
		return time.Time{}, heartbeat.ErrFetchGenesisTimeFromDb
	}

	var genesisTimeFromDb time.Time
	err = hs.marshalizer.Unmarshal(&genesisTimeFromDb, genesisTimeFromDbBytes)
	if err != nil {
		return time.Time{}, heartbeat.ErrUnmarshalGenesisTime
	}

	return genesisTimeFromDb, nil
}

// UpdateGenesisTime will update the saved genesis time and will log if the genesis time changed
func (hs *HeartbeatDbStorer) UpdateGenesisTime(genesisTime time.Time) error {
	if hs.storer.Has([]byte(genesisTimeDbEntry)) == nil { // if found, check for changes
		genesisTimeFromDb, err := hs.LoadGenesisTime()
		if err != nil {
			return err
		}

		if genesisTimeFromDb != genesisTime {
			log.Info(fmt.Sprintf("updated heartbeat's genesis time to %s", genesisTimeFromDb))
		}

		err = hs.saveGenesisTimeToDb(genesisTime)
		if err != nil {
			return err
		}
	}

	err := hs.saveGenesisTimeToDb(genesisTime)
	if err != nil {
		return err
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

// LoadHbmiDTO will return the HeartbeatDTO for the given public key from storage
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

// LoadKeys will return the keys saved in the storer, representing public keys of all peers the node is connected to
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

// SaveKeys will update the keys for all connected peers
func (hs *HeartbeatDbStorer) SaveKeys(peersSlice [][]byte) error {
	marshalizedFullPeersSlice, errMarsh := hs.marshalizer.Marshal(peersSlice)
	if errMarsh != nil {
		return errMarsh
	}

	return hs.storer.Put([]byte(peersKeysDbEntry), marshalizedFullPeersSlice)
}

// SavePubkeyData will add or update a HeartbeatDTO in the storer
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
		return errStore
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
