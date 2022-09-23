package storage_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/batch"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/heartbeat/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/storage"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
	storageStubs "github.com/ElrondNetwork/elrond-go/testscommon/storage"
	"github.com/stretchr/testify/assert"
)

func TestNewHeartbeatStorer_NilStorerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		nil,
		&mock.MarshallerStub{},
	)
	assert.Nil(t, hs)
	assert.Equal(t, heartbeat.ErrNilMonitorDb, err)
}

func TestNewHeartbeatStorer_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		&storageStubs.StorerStub{},
		nil,
	)
	assert.Nil(t, hs)
	assert.Equal(t, heartbeat.ErrNilMarshaller, err)
}

func TestNewHeartbeatStorer_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	hs, err := storage.NewHeartbeatDbStorer(
		&storageStubs.StorerStub{},
		&mock.MarshallerStub{},
	)
	assert.Nil(t, err)
	assert.False(t, check.IfNil(hs))
}

func TestHeartbeatDbStorer_LoadKeysEntryNotFoundShouldErr(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Nil(t, restoredKeys)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "not found"))
}

func TestHeartbeatDbStorer_LoadKeysUnmarshalInvalidShouldErr(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	keysBytes := []byte("invalid keys slice")
	_ = storer.Put([]byte("keys"), keysBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshallerMock{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Nil(t, restoredKeys)
	assert.NotNil(t, err)
}

func TestHeartbeatDbStorer_LoadKeysShouldWork(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	keys := [][]byte{[]byte("key1"), []byte("key2")}
	msr := &mock.MarshallerMock{}
	keysBytes, _ := msr.Marshal(&batch.Batch{Data: keys})
	_ = storer.Put([]byte("keys"), keysBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshallerMock{},
	)

	restoredKeys, err := hs.LoadKeys()
	assert.Nil(t, err)
	assert.Equal(t, keys, restoredKeys)
}

func TestHeartbeatDbStorer_SaveKeys(t *testing.T) {
	t.Parallel()

	keys := [][]byte{[]byte("key1"), []byte("key2")}
	hs, _ := storage.NewHeartbeatDbStorer(
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
	)

	err := hs.SaveKeys(keys)
	assert.Nil(t, err)

	restoredKeys, _ := hs.LoadKeys()
	assert.Equal(t, keys, restoredKeys)
}

func TestHeartbeatDbStorer_LoadGenesisTimeNotFoundInDbShouldErr(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
	)

	_, err := hs.LoadGenesisTime()
	assert.Equal(t, heartbeat.ErrFetchGenesisTimeFromDb, err)
}

func TestHeartbeatDbStorer_LoadGenesisUnmarshalIssueShouldErr(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	_ = storer.Put([]byte("genesisTime"), []byte("wrong genesis time"))

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		&mock.MarshallerMock{},
	)

	_, err := hs.LoadGenesisTime()
	assert.Equal(t, heartbeat.ErrUnmarshalGenesisTime, err)
}

func TestHeartbeatDbStorer_LoadGenesisTimeShouldWork(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	msr := &mock.MarshallerMock{}

	dbt := &data.DbTimeStamp{
		Timestamp: time.Now().UnixNano(),
	}
	expectedTime := time.Unix(0, dbt.Timestamp)

	genTimeBytes, _ := msr.Marshal(dbt)
	_ = storer.Put([]byte("genesisTime"), genTimeBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		msr,
	)

	recGenTime, err := hs.LoadGenesisTime()
	assert.Nil(t, err)
	assert.Equal(t, expectedTime.Second(), recGenTime.Second())
}

func TestHeartbeatDbStorer_UpdateGenesisTimeShouldFindAndReplace(t *testing.T) {
	t.Parallel()

	storer := genericMocks.NewStorerMock()
	msr := &mock.MarshallerMock{}

	dbt := &data.DbTimeStamp{
		Timestamp: time.Now().UnixNano(),
	}

	genTimeBytes, _ := msr.Marshal(dbt)
	_ = storer.Put([]byte("genesisTime"), genTimeBytes)

	hs, _ := storage.NewHeartbeatDbStorer(
		storer,
		msr,
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
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
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
		genericMocks.NewStorerMock(),
		&mock.MarshallerStub{
			MarshalHandler: func(obj interface{}) ([]byte, error) {
				return nil, expectedErr
			},
		},
	)

	hb := data.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Equal(t, expectedErr, err)
}

func TestHeartbeatDbSnorer_SavePubkeyDataPutNotSucceededShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("error putting")
	hs, _ := storage.NewHeartbeatDbStorer(
		&storageStubs.StorerStub{
			PutCalled: func(key, data []byte) error {
				return expectedErr
			},
		},
		&mock.MarshallerMock{},
	)

	hb := data.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Equal(t, expectedErr, err)
}

func TestHeartbeatDbSnorer_SavePubkeyDataPutShouldWork(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
	)

	hb := data.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	err := hs.SavePubkeyData([]byte("key1"), &hb)
	assert.Nil(t, err)
}

func TestHeartbeatDbStorer_LoadHeartBeatDTOShouldWork(t *testing.T) {
	t.Parallel()

	hs, _ := storage.NewHeartbeatDbStorer(
		genericMocks.NewStorerMock(),
		&mock.MarshallerMock{},
	)

	hb := data.HeartbeatDTO{
		NodeDisplayName: "test",
	}
	_ = hs.SavePubkeyData([]byte("key1"), &hb)

	hbmiDto, err := hs.LoadHeartBeatDTO("key1")
	assert.Nil(t, err)
	assert.Equal(t, hb.NodeDisplayName, hbmiDto.NodeDisplayName)
}
