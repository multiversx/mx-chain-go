package incomingHeader

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/sovereign"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	sovTests "github.com/multiversx/mx-chain-go/testscommon/sovereign"
)

func TestDepositEventProc_createEventData(t *testing.T) {
	t.Parallel()

	t.Run("empty transfer data", func(t *testing.T) {
		t.Parallel()

		inputEventData := []byte("data")

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(data []byte) (*sovereign.EventData, error) {
				require.Equal(t, inputEventData, data)

				return &sovereign.EventData{
					TransferData: nil,
				}, nil
			},
		}
		handler := &depositEventProc{
			marshaller:    args.Marshaller,
			hasher:        args.Hasher,
			dataCodec:     args.DataCodec,
			topicsChecker: args.TopicsChecker,
		}

		ret, err := handler.createEventData(inputEventData)
		require.Nil(t, err)
		require.Equal(t, &eventData{
			functionCallWithArgs: make([]byte, 0),
		}, ret)
	})

	t.Run("transfer data with function no args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{
					TransferData: &sovereign.TransferData{
						Function: func1,
					},
				}, nil
			},
		}
		handler := &depositEventProc{
			marshaller:    args.Marshaller,
			hasher:        args.Hasher,
			dataCodec:     args.DataCodec,
			topicsChecker: args.TopicsChecker,
		}

		ret, err := handler.createEventData([]byte(""))
		require.Nil(t, err)

		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)
		require.Equal(t, &eventData{
			functionCallWithArgs: expectedArgs,
		}, ret)
	})

	t.Run("transfer data with function and args", func(t *testing.T) {
		t.Parallel()

		func1 := []byte("func1")
		arg1 := []byte("arg1")
		arg2 := []byte("arg2")

		args := createArgs()
		args.DataCodec = &sovTests.DataCodecMock{
			DeserializeEventDataCalled: func(_ []byte) (*sovereign.EventData, error) {
				return &sovereign.EventData{
					TransferData: &sovereign.TransferData{
						Function: func1,
						Args:     [][]byte{arg1, arg2},
					},
				}, nil
			},
		}

		handler := &depositEventProc{
			marshaller:    args.Marshaller,
			hasher:        args.Hasher,
			dataCodec:     args.DataCodec,
			topicsChecker: args.TopicsChecker,
		}

		ret, err := handler.createEventData([]byte(""))
		require.Nil(t, err)

		expectedArgs := append([]byte("@"), hex.EncodeToString(func1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg1)...)
		expectedArgs = append(expectedArgs, "@"+hex.EncodeToString(arg2)...)
		require.Equal(t, &eventData{
			functionCallWithArgs: expectedArgs,
		}, ret)
	})
}

func TestDepositEventProc_createSCRData(t *testing.T) {
	t.Parallel()

	topicID := []byte("topicID")
	receiver := []byte("rcv")
	nft := []byte("nft")
	nonce := []byte("nonce")
	nftData := []byte("nftData")

	topics := [][]byte{
		topicID,
		receiver,
		nft,
		nonce,
		nftData,
	}

	args := createArgs()
	args.DataCodec = &sovTests.DataCodecMock{
		DeserializeTokenDataCalled: func(_ []byte) (*sovereign.EsdtTokenData, error) {
			return &sovereign.EsdtTokenData{
				TokenType: core.NonFungible,
				Royalties: big.NewInt(0),
			}, nil
		},
	}
	args.Marshaller = &marshallerMock.MarshalizerStub{
		MarshalCalled: func(_ interface{}) ([]byte, error) {
			return nftData, nil
		},
	}

	handler := &depositEventProc{
		marshaller:    args.Marshaller,
		hasher:        args.Hasher,
		dataCodec:     args.DataCodec,
		topicsChecker: args.TopicsChecker,
	}

	ret, err := handler.createSCRData(topics)
	require.Nil(t, err)

	expectedSCR := []byte(core.BuiltInFunctionMultiESDTNFTTransfer + "@01")
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nft)...)
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nonce)...)
	expectedSCR = append(expectedSCR, "@"+hex.EncodeToString(nftData)...)
	require.Equal(t, expectedSCR, ret)
}
