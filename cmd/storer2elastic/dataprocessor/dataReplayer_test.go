package dataprocessor_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/data"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/databasereader"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/dataprocessor"
	"github.com/ElrondNetwork/elrond-go/cmd/storer2elastic/mock"
	nodeConfig "github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/data/transaction"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/stretchr/testify/require"
)

func TestNewDataReplayer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() dataprocessor.DataReplayerArgs
		exError  error
	}{
		{
			name: "NilShardCoordinator",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.ShardCoordinator = nil
				return args
			},
			exError: dataprocessor.ErrNilShardCoordinator,
		},
		{
			name: "NilDatabaseReader",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.DatabaseReader = nil
				return args
			},
			exError: dataprocessor.ErrNilDatabaseReader,
		},
		{
			name: "NilMarshalizer",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.Marshalizer = nil
				return args
			},
			exError: dataprocessor.ErrNilMarshalizer,
		},
		{
			name: "NilHasher",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.Hasher = nil
				return args
			},
			exError: dataprocessor.ErrNilHasher,
		},
		{
			name: "NilUint64ByteSliceConverter",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.Uint64ByteSliceConverter = nil
				return args
			},
			exError: dataprocessor.ErrNilUint64ByteSliceConverter,
		},
		{
			name: "NilHeaderMarshalizer",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				args := getDataReplayArgs()
				args.HeaderMarshalizer = nil
				return args
			},
			exError: dataprocessor.ErrNilHeaderMarshalizer,
		},
		{
			name: "All arguments ok",
			argsFunc: func() dataprocessor.DataReplayerArgs {
				return getDataReplayArgs()
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dataprocessor.NewDataReplayer(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestDataReplayer_Range_NilHandlerShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataReplayArgs()
	dr, _ := dataprocessor.NewDataReplayer(args)

	err := dr.Range(nil)
	require.Equal(t, dataprocessor.ErrNilHandlerFunc, err)
}

func TestDataReplayer_Range_NoDatabaseFoundShouldErr(t *testing.T) {
	t.Parallel()

	args := getDataReplayArgs()
	dr, _ := dataprocessor.NewDataReplayer(args)

	handlerFunc := func(persistedData data.RoundPersistedData) bool {
		return true
	}
	err := dr.Range(handlerFunc)
	require.Equal(t, dataprocessor.ErrNoMetachainDatabase, err)
}

func TestDataReplayer_Range_EpochStartMetaBlockFoundForEpochStartKeyShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	shard0HdrHash := []byte("sh0Hdr")
	txMbHash := []byte("txMbHash")
	rwdTxMbHash := []byte("rwdTxMbHash")
	uTxMbHash := []byte("uTxMbHash")
	txHash := []byte("txHash")
	rwdTxHash := []byte("rwdTx")
	uTxHash := []byte("uTxHash")

	epStMb := &block.MetaBlock{Nonce: 0, Epoch: 0, ShardInfo: []block.ShardData{{HeaderHash: shard0HdrHash}}}
	epStMbBytes, _ := marshalizer.Marshal(epStMb)
	metaBlocksPersister := mock.NewPersisterMock()
	_ = metaBlocksPersister.Put([]byte("epochStartBlock_0"), epStMbBytes)

	sh0Hdr := &block.Header{
		Round: 0,
		MiniBlockHeaders: []block.MiniBlockHeader{
			{Type: block.TxBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: 0, Hash: txMbHash},
			{Type: block.RewardsBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: 0, Hash: rwdTxMbHash},
			{Type: block.SmartContractResultBlock, SenderShardID: core.MetachainShardId, ReceiverShardID: 0, Hash: uTxMbHash},
		},
	}
	sh0HdrBytes, _ := marshalizer.Marshal(sh0Hdr)
	shardHeadersPersister := mock.NewPersisterMock()
	_ = shardHeadersPersister.Put(shard0HdrHash, sh0HdrBytes)

	miniBlocksPersister := mock.NewPersisterMock()
	mbHdr := &block.MiniBlock{Type: block.TxBlock, TxHashes: [][]byte{txHash}}
	mbHdrBytes, _ := marshalizer.Marshal(mbHdr)
	_ = miniBlocksPersister.Put(txMbHash, mbHdrBytes)

	rwdMbHdr := &block.MiniBlock{Type: block.RewardsBlock, TxHashes: [][]byte{rwdTxHash}}
	rwdMbHdrBytes, _ := marshalizer.Marshal(rwdMbHdr)
	_ = miniBlocksPersister.Put(rwdTxMbHash, rwdMbHdrBytes)

	uMbHdr := &block.MiniBlock{Type: block.SmartContractResultBlock, TxHashes: [][]byte{uTxHash}}
	uMbHdrBytes, _ := marshalizer.Marshal(uMbHdr)
	_ = miniBlocksPersister.Put(uTxMbHash, uMbHdrBytes)

	txPersister := mock.NewPersisterMock()
	tx := &transaction.Transaction{Nonce: 37}
	txBytes, _ := marshalizer.Marshal(tx)

	rwdTx := rewardTx.RewardTx{Round: 37}
	rwdTxBytes, _ := marshalizer.Marshal(rwdTx)

	uTx := smartContractResult.SmartContractResult{Nonce: 37}
	uTxBytes, _ := marshalizer.Marshal(uTx)

	_ = txPersister.Put(txHash, txBytes)
	_ = txPersister.Put(rwdTxHash, rwdTxBytes)
	_ = txPersister.Put(uTxHash, uTxBytes)

	args := getDataReplayArgs()
	args.DatabaseReader = &mock.DatabaseReaderStub{
		GetDatabaseInfoCalled: func() ([]*databasereader.DatabaseInfo, error) {
			return getDbInfoMetaAndOneShard(), nil
		},
		GetStaticDatabaseInfoCalled: func() ([]*databasereader.DatabaseInfo, error) {
			return getDbInfoMetaAndOneShard(), nil
		},
		GetHeadersCalled: nil,
		LoadPersisterCalled: func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
			switch unit {
			case "MetaBlock":
				return metaBlocksPersister, nil
			case "BlockHeaders":
				return shardHeadersPersister, nil
			case "MiniBlocks":
				return miniBlocksPersister, nil
			case "Transactions", "RewardTransactions", "UnsignedTransactions":
				return txPersister, nil
			}

			return mock.NewPersisterMock(), nil
		},
		LoadStaticPersisterCalled: func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
			return mock.NewPersisterMock(), nil
		},
	}
	args.ShardCoordinator = &mock.ShardCoordinatorMock{ShardID: 0, NumOfShards: 1}
	realHeaderMarshalizer, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	args.HeaderMarshalizer = realHeaderMarshalizer
	dr, _ := dataprocessor.NewDataReplayer(args)

	handlerFunc := func(persistedData data.RoundPersistedData) bool {
		return true
	}
	_ = dr.Range(handlerFunc)
}

func TestDataReplayer_Range_EpochStartMetaBlockFoundInNonceHashStorageShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}

	metaBlckHash := []byte("epochStart")

	epStMb := &block.MetaBlock{Nonce: 0, Epoch: 0, ShardInfo: []block.ShardData{{HeaderHash: []byte("tada")}}}
	epStMbBytes, _ := marshalizer.Marshal(epStMb)
	metaBlocksPersister := mock.NewPersisterMock()
	err := metaBlocksPersister.Put(metaBlckHash, epStMbBytes)
	require.NoError(t, err)

	uint64ByteSliceConv := &mock.Uint64ByteSliceConverterMock{}
	nonce0Bytes := uint64ByteSliceConv.ToByteSlice(0)
	hdrHashPersister := mock.NewPersisterMock()
	_ = hdrHashPersister.Put(nonce0Bytes, metaBlckHash)

	args := getDataReplayArgs()
	args.DatabaseReader = &mock.DatabaseReaderStub{
		GetDatabaseInfoCalled: func() ([]*databasereader.DatabaseInfo, error) {
			return getDbInfoMetaAndOneShard(), nil
		},
		GetStaticDatabaseInfoCalled: func() ([]*databasereader.DatabaseInfo, error) {
			return getDbInfoMetaAndOneShard(), nil
		},
		GetHeadersCalled: nil,
		LoadPersisterCalled: func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
			if unit == "MetaBlock" {
				return metaBlocksPersister, nil
			}

			return mock.NewPersisterMock(), nil
		},
		LoadStaticPersisterCalled: func(dbInfo *databasereader.DatabaseInfo, unit string) (storage.Persister, error) {
			return hdrHashPersister, nil
		},
	}
	args.ShardCoordinator = &mock.ShardCoordinatorMock{ShardID: 0, NumOfShards: 1}
	realHeaderMarshalizer, _ := databasereader.NewHeaderMarshalizer(marshalizer)
	args.HeaderMarshalizer = realHeaderMarshalizer
	dr, _ := dataprocessor.NewDataReplayer(args)

	handlerFunc := func(persistedData data.RoundPersistedData) bool {
		return true
	}
	_ = dr.Range(handlerFunc)
}

func getDataReplayArgs() dataprocessor.DataReplayerArgs {
	return dataprocessor.DataReplayerArgs{
		GeneralConfig: nodeConfig.Config{
			MetaBlockStorage:        nodeConfig.StorageConfig{DB: nodeConfig.DBConfig{FilePath: "MetaBlock"}},
			MetaHdrNonceHashStorage: nodeConfig.StorageConfig{DB: nodeConfig.DBConfig{FilePath: "MetaHdrHashNonce"}},
		},
		DatabaseReader:           &mock.DatabaseReaderStub{},
		ShardCoordinator:         &mock.ShardCoordinatorMock{},
		Marshalizer:              &mock.MarshalizerMock{},
		Hasher:                   &mock.HasherMock{},
		Uint64ByteSliceConverter: &mock.Uint64ByteSliceConverterMock{},
		HeaderMarshalizer:        &mock.HeaderMarshalizerStub{},
	}
}

func getDbInfoMetaAndOneShard() []*databasereader.DatabaseInfo {
	return []*databasereader.DatabaseInfo{
		{
			Epoch: 0,
			Shard: core.MetachainShardId,
		},
		{
			Epoch: 0,
			Shard: 0,
		},
	}
}
