package transactionAPI

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newFieldsHandler(t *testing.T) {
	t.Parallel()

	fh := newFieldsHandler("")
	require.Equal(t, fieldsHandler{make(map[string]struct{})}, fh)

	providedFields := "nOnCe,sender,receiver,gasLimit,GASprice,receiverusername,data,value,signature,guardian,guardiansignature,sendershard,receivershard"
	splitFields := strings.Split(providedFields, separator)
	fh = newFieldsHandler(providedFields)
	for _, field := range splitFields {
		require.True(t, fh.IsFieldSet(field), fmt.Sprintf("field %s is not set", field))
	}

	fh = newFieldsHandler("*")
	for _, field := range splitFields {
		require.True(t, fh.IsFieldSet(field))
	}
}
