package receipts

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/hashing/blake2b"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/genericMocks"
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
			Marshaller: testscommon.MarshalizerMock{},
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
			Marshaller: testscommon.MarshalizerMock{},
			Hasher:     &testscommon.HasherStub{},
			Store:      nil,
		}

		repository, err := NewReceiptsRepository(arguments)
		require.ErrorIs(t, err, errCannotCreateReceiptsRepository)
		require.ErrorContains(t, err, core.ErrNilStore.Error())
		require.True(t, check.IfNil(repository))
	})

	t.Run("no error", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: testscommon.MarshalizerMock{},
			Hasher:     &testscommon.HasherStub{},
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.Nil(t, err)
		require.False(t, check.IfNil(repository))
	})

	t.Run("emptyReceiptsHash is set", func(t *testing.T) {
		arguments := ArgsNewReceiptsRepository{
			Marshaller: &marshal.GogoProtoMarshalizer{},
			Hasher:     blake2b.NewBlake2b(),
			Store:      genericMocks.NewChainStorerMock(0),
		}

		repository, err := NewReceiptsRepository(arguments)
		require.Nil(t, err)
		require.Equal(t, "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8", hex.EncodeToString(repository.emptyReceiptsHash))
	})
}

func TestReceiptsRepository_SaveReceipts(t *testing.T) {
	marshaller := &marshal.GogoProtoMarshalizer{}
	hasher := blake2b.NewBlake2b()
	emptyReceiptsHash, _ := createEmptyReceiptsHash(marshaller, hasher)
	nonEmptyReceiptsHash := []byte("receipts-hash")
	receiptsHolder := &ReceiptsHolder{Miniblocks: []*block.MiniBlock{{Type: block.SmartContractResultBlock}}}
	receiptsBytes, _ := marshalReceiptsHolder(receiptsHolder, marshaller)
	headerHash := []byte("header-hash")

	t.Run("when receipts hash is for non-empty receipts", func(t *testing.T) {
		store := genericMocks.NewChainStorerMock(0)

		repository, _ := NewReceiptsRepository(ArgsNewReceiptsRepository{
			Marshaller: marshaller,
			Hasher:     hasher,
			Store:      store,
		})

		repository.SaveReceipts(
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

		repository.SaveReceipts(
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

func TestReceiptsRepository_LoadReceipts(t *testing.T) {
	// TBD
}
