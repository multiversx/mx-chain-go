package transactionAPI

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newParametersHandler(t *testing.T) {
	t.Parallel()

	ph := newParametersHandler("")
	require.Equal(t, parametersHandler{}, ph)

	ph = newParametersHandler("hash,nonce,sender,receiver,gaslimit,gasprice")
	expectedPH := parametersHandler{
		HasHash:     true,
		HasNonce:    true,
		HasSender:   true,
		HasReceiver: true,
		HasGasLimit: true,
		HasGasPrice: true,
	}
	require.Equal(t, expectedPH, ph)
}
