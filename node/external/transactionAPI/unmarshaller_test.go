package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-go/process/mock"
	"github.com/stretchr/testify/require"
)

func TestGetEncodedAddress(t *testing.T) {
	t.Parallel()

	address := []byte("address")
	expectedEncodedAddr := "encoded"
	txUnmarshalledHandler := &txUnmarshaller{
		addressPubKeyConverter: &mock.PubkeyConverterStub{
			LenCalled: func() int {
				return len(address)
			},
			EncodeCalled: func(pkBytes []byte) string {
				require.Equal(t, pkBytes, address)
				return expectedEncodedAddr
			},
		},
	}

	encodedAddr := txUnmarshalledHandler.getEncodedAddress(address)
	require.Equal(t, expectedEncodedAddr, encodedAddr)

	encodedAddr = txUnmarshalledHandler.getEncodedAddress([]byte("adr"))
	require.Empty(t, encodedAddr)
}

func TestBigIntToStr(t *testing.T) {
	t.Parallel()

	val := bigIntToStr(big.NewInt(123))
	require.Equal(t, "123", val)

	val = bigIntToStr(nil)
	require.Empty(t, val)
}
