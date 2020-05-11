package storage

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-logger"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/batch"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.GetOrCreate("heartbeat/storage")

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
	if check.IfNil(storer) {
		return nil, heartbeat.ErrNilMonitorDb
	}
	if check.IfNil(marshalizer) {
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

	dbts := &data.DbTimeStamp{}
	err = hs.marshalizer.Unmarshal(dbts, genesisTimeFromDbBytes)
	if err != nil {
		return time.Time{}, heartbeat.ErrUnmarshalGenesisTime
	}

	genesisTimeFromDb := time.Unix(0, dbts.Timestamp)

	return genesisTimeFromDb, nil
}

// UpdateGenesisTime will update the saved genesis time and will log if the genesis time changed
func (hs *HeartbeatDbStorer) UpdateGenesisTime(genesisTime time.Time) error {
	genesisTimeFromDb, err := hs.LoadGenesisTime()
	if err != nil && err != heartbeat.ErrFetchGenesisTimeFromDb {
		return err
	}

	err = hs.saveGenesisTimeToDb(genesisTime)
	if err != nil {
		return err
	}

	if genesisTimeFromDb != genesisTime {
		log.Debug("updated heartbeat's genesis time",
			"time [s]", genesisTimeFromDb)
	}

	return nil
}

func (hs *HeartbeatDbStorer) saveGenesisTimeToDb(genesisTime time.Time) error {
	dbts := &data.DbTimeStamp{
		Timestamp: genesisTime.UnixNano(),
	}

	genesisTimeBytes, err := hs.marshalizer.Marshal(dbts)
	if err != nil {
		return heartbeat.ErrMarshalGenesisTime
	}

	err = hs.storer.Put([]byte(genesisTimeDbEntry), genesisTimeBytes)
	if err != nil {
		return heartbeat.ErrStoreGenesisTimeToDb
	}

	return nil
}

// LoadHeartBeatDTO will return the HeartbeatDTO for the given public key from storage
func (hs *HeartbeatDbStorer) LoadHeartBeatDTO(pubKey string) (*data.HeartbeatDTO, error) {
	pkbytes := []byte(pubKey)

	hbFromDB, err := hs.storer.Get(pkbytes)
	if err != nil {
		return nil, err
	}

	heartbeatDto := data.HeartbeatDTO{}
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

	b := &batch.Batch{}
	err = hs.marshalizer.Unmarshal(b, allKeysBytes)
	if err != nil {
		return nil, err
	}

	return b.Data, nil
}

// SaveKeys will update the keys for all connected peers
func (hs *HeartbeatDbStorer) SaveKeys(peersSlice [][]byte) error {
	marshalizedFullPeersSlice, errMarsh := hs.marshalizer.Marshal(&batch.Batch{Data: peersSlice})
	if errMarsh != nil {
		return errMarsh
	}

	return hs.storer.Put([]byte(peersKeysDbEntry), marshalizedFullPeersSlice)
}

// SavePubkeyData will add or update a HeartbeatDTO in the storer
func (hs *HeartbeatDbStorer) SavePubkeyData(
	pubkey []byte,
	heartbeat *data.HeartbeatDTO,
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
	return hs == nil
}
