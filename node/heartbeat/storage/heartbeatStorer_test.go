package storage_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go/node/heartbeat"
	"github.com/ElrondNetwork/elrond-go/node/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewHeartbeatStorer_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		nil,
		&mock.MarshalizerMock{},
	)
	assert.Nil(t, hs)
	assert.Equal(t, heartbeat.ErrNilMonitorDb, err)
}

func TestNewHeartbeatStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		&mock.StorerStub{},
		nil,
	)
	assert.Nil(t, hs)
	assert.Equal(t, heartbeat.ErrNilMarshalizer, err)
}

func TestNewHeartbeatStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		&mock.StorerStub{},
		&mock.MarshalizerMock{},
	)
	assert.NotNil(t, hs)
	assert.Nil(t, err)
}

func TestHeartbeatDbStorer_LoadKeysEntryNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Nil(t, restoredKeys)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestHeartbeatDbStorer_LoadKeysUnmarshalInvalidShouldErr(t *testing.T) {
	t.Parallel()

	storer := mock.NewStorerMock()
	keysBytes := []byte("invalid keys slice")
	_ = storer.Put([]byte("keys"), keysBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshalizerFake{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Nil(t, restoredKeys)
	assert.NotNil(t, err)
}

func TestHeartbeatDbStorer_LoadKeysShouldWork(t *testing.T) {
	t.Parallel()

	storer := mock.NewStorerMock()
	keys := [][]byte{[]byte("key1"), []byte("key2")}
	keysBytes, _ := json.Marshal(keys)
	_ = storer.Put([]byte("keys"), keysBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshalizerFake{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Equal(t, keys, restoredKeys)
	assert.Nil(t, err)
}

func TestHeartbeatDbStorer_SaveKeys(t *testing.T) {
	t.Parallel()

	keys := [][]byte{[]byte("key1"), []byte("key2")}
	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	err := hs.SaveKeys(keys)
	assert.Nil(t, err)

	restoredKeys, _ := hs.LoadKeys()
	assert.Equal(t, keys, restoredKeys)
}

func TestHeartbeatDbStorer_LoadGenesisTimeNotFoundInDbShouldErr(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	_, err := hs.LoadGenesisTime()
	assert.Equal(t, heartbeat.ErrFetchGenesisTimeFromDb, err)
}

func TestHeartbeatDbStorer_LoadGenesisUnmarshalIssueShouldErr(t *testing.T) {
	t.Parallel()

	storer := mock.NewStorerMock()
	_ = storer.Put([]byte("genesisTime"), []byte("wrong genesis time"))

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshalizerFake{},
	)

	_, err := hs.LoadGenesisTime()
	assert.Equal(t, heartbeat.ErrUnmarshalGenesisTime, err)
}

func TestHeartbeatDbStorer_LoadGenesisTimeShouldWork(t *testing.T) {
	t.Parallel()

	expectedTime := time.Now()
	storer := mock.NewStorerMock()
	genTimeBytes, _ := json.Marshal(expectedTime)
	_ = storer.Put([]byte("genesisTime"), genTimeBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshalizerFake{},
	)

	recGenTime, err := hs.LoadGenesisTime()
	assert.Nil(t, err)
	assert.Equal(t, expectedTime.Second(), recGenTime.Second())
}

func TestHeartbeatDbStorer_UpdateGenesisTimeShouldFindAndReplace(t *testing.T) {
	t.Parallel()

	expectedTime := time.Now()
	storer := mock.NewStorerMock()
	genTimeBytes, _ := json.Marshal(expectedTime)
	_ = storer.Put([]byte("genesisTime"), genTimeBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshalizerFake{},
	)

	newGenesisTime := time.Now()
	err := hs.UpdateGenesisTime(newGenesisTime)
	assert.Nil(t, err)

	recGenTime, _ := hs.LoadGenesisTime()
	assert.Equal(t, newGenesisTime.Second(), recGenTime.Second())
}

func TestHeartbeatDbStorer_UpdateGenesisTimeShouldAddNewEntry(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	genesisTime := time.Now()
	err := hs.UpdateGenesisTime(genesisTime)
	assert.Nil(t, err)

	recGenTime, _ := hs.LoadGenesisTime()
	assert.Equal(t, genesisTime.Second(), recGenTime.Second())
}

func TestHeartbeatDbSnorer_SavePubkeyDataDataMarshalNotSucceededShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error marshal")
	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerMock{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		},
	)

	hb := heartbeat.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Equal(t, expectedErr, err)
}

func TestHeartbeatDbSnorer_SavePubkeyDataPutNotSucceededShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error putting")
	hs, _ := storage.NewHeartbeatDbStorer(
		&mock.StorerStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		},
		&mock.MarshalizerFake{},
	)

	hb := heartbeat.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Equal(t, expectedErr, err)
}

func TestHeartbeatDbSnorer_SavePubkeyDataPutShouldWork(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	hb := heartbeat.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Nil(t, err)
}

func TestHeartbeatDbStorer_LoadHbmiDTOShouldWork(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		mock.NewStorerMock(),
		&mock.MarshalizerFake{},
	)

	hb := heartbeat.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	_ = hs.SavePubkeyData([]byte("key1"), &hb)

	hbmiDto, err := hs.LoadHbmiDTO("key1")
	assert.Nil(t, err)
	assert.Equal(t, hb.NodeDisplayName, hbmiDto.NodeDisplayName)
}
