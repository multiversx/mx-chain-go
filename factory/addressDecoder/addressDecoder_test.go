package addressDecoder_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	chainError "github.com/multiversx/mx-chain-go/errors"
	"github.com/multiversx/mx-chain-go/factory/addressDecoder"
	"github.com/multiversx/mx-chain-go/testscommon"
)

func Test_DecodeAddressesError(t *testing.T) {
	t.Parallel()

	t.Run("nil pkConverter", func(t *testing.T) {
		t.Parallel()

		addresses, err := addressDecoder.DecodeAddresses(nil, make([]string, 0))
		assert.Nil(t, addresses)
		assert.Equal(t, chainError.ErrNilPubKeyConverter, err)
	})
	t.Run("decode error", func(t *testing.T) {
		t.Parallel()
		pkError := errors.New("pkerror")
		pkConverter := &testscommon.PubkeyConverterStub{
			DecodeCalled: func(humanReadable string) ([]byte, error) {
				return nil, pkError
			}}

		stringAddresses := []string{"addr1"}
		addresses, err := addressDecoder.DecodeAddresses(pkConverter, stringAddresses)
		assert.Nil(t, addresses)
		assert.Equal(t, pkError, err)
	})
}

func Test_DecodeAddressesShouldWork(t *testing.T) {
	t.Parallel()

	decodedAddr1 := []byte("decodedAddr1")
	decodedAddr2 := []byte("decodedAddr2")

	decodeMap := make(map[string][]byte)
	decodeMap["addr1"] = decodedAddr1
	decodeMap["addr2"] = decodedAddr2

	pkConverter := &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return decodeMap[humanReadable], nil
		}}

	stringAddresses := []string{"addr1", "addr2"}
	addresses, err := addressDecoder.DecodeAddresses(pkConverter, stringAddresses)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(addresses))
	assert.Equal(t, decodedAddr1, addresses[0])
	assert.Equal(t, decodedAddr2, addresses[1])
}
