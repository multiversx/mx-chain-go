package receipts

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/hashing/blake2b"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/common/holders"
	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/storage"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genericMocks"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	testsCommonStorage "github.com/multiversx/mx-chain-go/testscommon/storage"
	"github.com/stretchr/testify/require"
)

func TestNewReceiptsRepository(t *testing.T) {
	t.Run("NilMarshaller", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: nil,
			Hasher:     &testscommon.HasherStub{},
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.ErrorIs(t, err, errCannotCreateReceiptsRepository)
		require.ErrorContains(t, err, core.ErrNilMarshalizer.Error())
		require.True(t, check.IfNil(repository))
	})

	t.Run("NilHasher", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: marshallerMock.MarshalizerMock{},
			Hasher:     nil,
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.ErrorIs(t, err, errCannotCreateReceiptsRepository)
		require.ErrorContains(t, err, core.ErrNilHasher.Error())
		require.True(t, check.IfNil(repository))
	})

	t.Run("NilStorer", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: marshallerMock.MarshalizerMock{},
			Hasher:     &testscommon.HasherStub{},
			Store:      nil,
		}

		repository, err := NewReceiptsRepository(arguments)
		require.ErrorIs(t, err, errCannotCreateReceiptsRepository)
		require.ErrorContains(t, err, core.ErrNilStore.Error())
		require.True(t, check.IfNil(repository))
	})

	t.Run("storer not found", func(t *testing.T) {
		expectedErr := errors.New("expected error")
		arguments := ArgsNewReceiptsRepository{
			Marshaller: marshallerMock.MarshalizerMock{},
			Hasher:     &testscommon.HasherStub{},
			Store: &testsCommonStorage.ChainStorerStub{
				GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
					return nil, expectedErr
				},
			},
		}

		repository, err := NewReceiptsRepository(arguments)
		require.ErrorIs(t, err, errCannotCreateReceiptsRepository)
		require.ErrorContains(t, err, expectedErr.Error())
		require.True(t, check.IfNil(repository))
	})

	t.Run("no error", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: marshallerMock.MarshalizerMock{},
			Hasher:     &testscommon.HasherStub{},
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.Nil(t, err)
		require.False(t, check.IfNil(repository))
	})

	t.Run("emptyReceiptsHash is cached", func(t *testing.T) {
		marshaller := &marshal.GogoProtoMarshalizer{}
		hasher := blake2b.NewBlake2b()
		emptyReceiptsHash, _ := createEmptyReceiptsHash(marshaller, hasher)

		arguments := ArgsNewReceiptsRepository{
			Marshaller: marshaller,
			Hasher:     hasher,
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.Nil(t, err)
		require.Equal(t, emptyReceiptsHash, repository.emptyReceiptsHash)
	})
}

func TestReceiptsRepository_SaveReceipts(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()
	emptyReceiptsHash, _ := createEmptyReceiptsHash(marshaller, hasher)
	nonEmptyReceiptsHash := []byte("non-empty-receipts-hash")
	receiptsHolder := holders.NewReceiptsHolder([]*block.MiniBlock{{Type: block.SmartContractResultBlock}})
	receiptsBytes, _ := marshalReceiptsHolder(receiptsHolder, marshaller)
	headerHash := []byte("header-hash")

	t.Run("when receipts hash is for non-empty receipts", func(t *testing.T) {
		store := genericMocks.NewChainStorerMock(0)

		repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
			Marshaller: marshaller,
			Hasher:     hasher,
			Store:      store,
		})

		_ = repository.SaveReceipts(
			receiptsHolder,
			&block.Header{ReceiptsHash: nonEmptyReceiptsHash},
			headerHash,
		)

		// Saved at key = nonEmptyReceiptsHash (header.GetReceiptsHash())
		actualData, err := store.Get(dataRetriever.ReceiptsUnit, nonEmptyReceiptsHash)
		require.Nil(t, err)
		require.Equal(t, receiptsBytes, actualData)

		// Not saved at key = headerHash
		_, err = store.Get(dataRetriever.ReceiptsUnit, headerHash)
		require.NotNil(t, err)
	})

	t.Run("when receipts hash is for empty receipts", func(t *testing.T) {
		store := genericMocks.NewChainStorerMock(0)

		repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
			Marshaller: marshaller,
			Hasher:     hasher,
			Store:      store,
		})

		_ = repository.SaveReceipts(
			receiptsHolder,
			&block.Header{ReceiptsHash: emptyReceiptsHash},
			headerHash,
		)

		// Saved at key = headerHash
		actualData, err := store.Get(dataRetriever.ReceiptsUnit, headerHash)
		require.Nil(t, err)
		require.Equal(t, receiptsBytes, actualData)

		// Not saved at key = emptyReceiptsHash
		_, err = store.Get(dataRetriever.ReceiptsUnit, emptyReceiptsHash)
		require.NotNil(t, err)
	})
}

func TestReceiptsRepository_SaveReceiptsShouldErrWhenPassingBadInput(t *testing.T) {
	repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
		Marshaller: &marshal.GogoProtoMarshalizer{},
		Hasher:     blake2b.NewBlake2b(),
		Store:      genericMocks.NewChainStorerMock(0),
	})

	t.Run("when receipts holder is nil", func(t *testing.T) {
		err := repository.SaveReceipts(nil, &block.Header{}, []byte("aaaa"))
		require.ErrorIs(t, err, errNilReceiptsHolder)
	})

	t.Run("when block header is nil", func(t *testing.T) {
		err := repository.SaveReceipts(holders.NewReceiptsHolder([]*block.MiniBlock{}), nil, []byte("aaaa"))
		require.ErrorIs(t, err, errNilBlockHeader)
	})

	t.Run("when block hash is empty", func(t *testing.T) {
		err := repository.SaveReceipts(holders.NewReceiptsHolder([]*block.MiniBlock{}), &block.Header{}, nil)
		require.ErrorIs(t, err, errEmptyBlockHash)
	})
}

func TestReceiptsRepository_LoadReceipts(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()
	store := genericMocks.NewChainStorerMock(0)
	emptyReceiptsHash, _ := createEmptyReceiptsHash(marshaller, hasher)
	nonEmptyReceiptsHash := []byte("non-empty-receipts-hash")
	headerHash := []byte("header-hash")

	repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
		Marshaller: marshaller,
		Hasher:     hasher,
		Store:      store,
	})

	receiptsAtKeyHeaderHash := holders.NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 42}})
	receiptsAtKeyHeaderHashBytes, _ := marshalReceiptsHolder(receiptsAtKeyHeaderHash, marshaller)
	_ = store.Put(dataRetriever.ReceiptsUnit, headerHash, receiptsAtKeyHeaderHashBytes)

	receiptsAtKeyReceiptsHash := holders.NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 43}})
	receiptsAtKeyReceiptsHashBytes, _ := marshalReceiptsHolder(receiptsAtKeyReceiptsHash, marshaller)
	_ = store.Put(dataRetriever.ReceiptsUnit, nonEmptyReceiptsHash, receiptsAtKeyReceiptsHashBytes)

	t.Run("when header.GetReceiptsHash() == emptyReceiptsHash", func(t *testing.T) {
		loaded, err := repository.LoadReceipts(&block.Header{ReceiptsHash: emptyReceiptsHash}, headerHash)
		require.Nil(t, err)
		require.Equal(t, receiptsAtKeyHeaderHash, loaded)
	})

	t.Run("when header.GetReceiptsHash() != emptyReceiptsHash", func(t *testing.T) {
		loaded, err := repository.LoadReceipts(&block.Header{ReceiptsHash: nonEmptyReceiptsHash}, headerHash)
		require.Nil(t, err)
		require.Equal(t, receiptsAtKeyReceiptsHash, loaded)
	})

	t.Run("when no receipts for given header", func(t *testing.T) {
		loadedHolder, err := repository.LoadReceipts(&block.Header{ReceiptsHash: emptyReceiptsHash}, []byte("abba"))
		require.Nil(t, err)
		require.Equal(t, createEmptyReceiptsHolder(), loadedHolder)
	})
}

func TestReceiptsRepository_NoPanicOnSaveOrLoadWhenBadStorage(t *testing.T) {
	store := &testsCommonStorage.ChainStorerStub{
		GetStorerCalled: func(unitType dataRetriever.UnitType) (storage.Storer, error) {
			return &testsCommonStorage.StorerStub{
				PutCalled: func(key, data []byte) error {
					return errors.New("bad")
				},
				GetFromEpochCalled: func(key []byte, epoch uint32) ([]byte, error) {
					return nil, errors.New("bad")
				},
			}, nil
		},
	}

	repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
		Marshaller: marshallerMock.MarshalizerMock{},
		Hasher:     &testscommon.HasherStub{},
		Store:      store,
	})

	t.Run("save in bad storage", func(t *testing.T) {
		holder := holders.NewReceiptsHolder([]*block.MiniBlock{{SenderShardID: 42}})
		header := &block.Header{ReceiptsHash: []byte("aaaa")}
		err := repository.SaveReceipts(holder, header, []byte("bbbb"))
		require.NotNil(t, err)
		require.ErrorIs(t, err, errCannotSaveReceipts)
	})

	t.Run("load from bad storage", func(t *testing.T) {
		header := &block.Header{ReceiptsHash: []byte("aaaa")}
		loaded, err := repository.LoadReceipts(header, []byte("bbbb"))
		require.NotNil(t, err)
		require.ErrorIs(t, err, errCannotLoadReceipts)
		require.Nil(t, loaded)
	})
}

func TestReceiptsRepository_DecideStorageKey(t *testing.T) {
	repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
		Marshaller: marshallerMock.MarshalizerMock{},
		Hasher:     &testscommon.HasherStub{},
		Store:      genericMocks.NewChainStorerMock(0),
	})

	repository.emptyReceiptsHash = []byte("empty-receipts-hash")

	t.Run("when receipts hash is for non-empty receipts", func(t *testing.T) {
		storageKey := repository.decideStorageKey([]byte("non-empty-receipts-hash"), []byte("header-hash"))
		require.Equal(t, []byte("non-empty-receipts-hash"), storageKey)
	})

	t.Run("when receipts hash is for empty receipts", func(t *testing.T) {
		storageKey := repository.decideStorageKey([]byte("empty-receipts-hash"), []byte("header-hash"))
		require.Equal(t, []byte("header-hash"), storageKey)
	})
}

func TestCreateEmptyReceiptsHash(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()

	emptyReceiptsHash, err := createEmptyReceiptsHash(marshaller, hasher)
	require.Nil(t, err)
	require.Equal(t, "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8", hex.EncodeToString(emptyReceiptsHash))
}
