package transactionAPI

import (
	"math/big"
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestGetEncodedAddress(t *testing.T) {
	t.Parallel()

	address := []byte("12345678901234567890123456789012")
	expectedEncodedAddr := "erd1xyerxdp4xcmnswfsxyerxdp4xcmnswfsxyerxdp4xcmnswfsxyeqlrqt99"
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
	require.Nil(t, err)
	require.Equal(t, expectedEncodedAddr, encodedAddr)

	encodedAddr, err = txUnmarshalledHandler.getEncodedAddress([]byte("abc"))
	require.Empty(t, encodedAddr)
	require.Error(t, ErrEncodeAddress, err)
}

func TestBigIntToStr(t *testing.T) {
	t.Parallel()

	val := bigIntToStr(big.NewInt(123))
	require.Equal(t, "123", val)

	val = bigIntToStr(nil)
	require.Empty(t, val)
}
