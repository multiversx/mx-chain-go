package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestGetEncodedAddress(t *testing.T) {
	t.Parallel()

	address := []byte("address")
	expectedEncodedAddr := "encoded"
	txUnmarshalledHandler := &txUnmarshaller{
		addressPubKeyConverter: &testscommon.PubkeyConverterStub{
			LenCalled: func() int {
				return len(address)
			},
			EncodeCalled: func(pkBytes []byte) (string, error) {
				require.Equal(t, pkBytes, address)
				return expectedEncodedAddr, nil
			},
		},
	}

	encodedAddr, err := txUnmarshalledHandler.getEncodedAddress(address)
	require.Equal(t, expectedEncodedAddr, encodedAddr)
	require.Nil(t, err)

	encodedAddr, err = txUnmarshalledHandler.getEncodedAddress([]byte("adr"))
	require.Empty(t, encodedAddr)
	require.Nil(t, err)
}

func TestBigIntToStr(t *testing.T) {
	t.Parallel()

	val := bigIntToStr(big.NewInt(123))
	require.Equal(t, "123", val)

	val = bigIntToStr(nil)
	require.Empty(t, val)
}
