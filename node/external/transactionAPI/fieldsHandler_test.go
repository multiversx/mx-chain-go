package transactionAPI

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newFieldsHandler(t *testing.T) {
	t.Parallel()

	fh := newFieldsHandler("")
	require.Equal(t, fieldsHandler{}, fh)

	fh = newFieldsHandler("nOnCe,sender,receiver,gasLimit,GASprice,receiverusername,data,value")
	expectedPH := fieldsHandler{
		HasNonce:       true,
		HasSender:      true,
		HasReceiver:    true,
		HasGasLimit:    true,
		HasGasPrice:    true,
		HasRcvUsername: true,
		HasData:        true,
		HasValue:       true,
	}
	require.Equal(t, expectedPH, fh)
}
