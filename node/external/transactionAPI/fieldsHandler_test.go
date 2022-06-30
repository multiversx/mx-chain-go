package transactionAPI

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newFieldsHandler(t *testing.T) {
	t.Parallel()

	fh := newFieldsHandler("")
	require.Equal(t, fieldsHandler{}, fh)

	fh = newFieldsHandler("nonce,sender,receiver,gaslimit,gasprice")
	expectedPH := fieldsHandler{
		HasNonce:    true,
		HasSender:   true,
		HasReceiver: true,
		HasGasLimit: true,
		HasGasPrice: true,
	}
	require.Equal(t, expectedPH, fh)
}
