package smartContract

import (
	"encoding/hex"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestNewAtArgumentParser(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	args, err := parser.GetArguments()
	assert.Nil(t, args)
	assert.Equal(t, process.ErrNilArguments, err)

	code, err := parser.GetCode()
	assert.Nil(t, code)
	assert.Equal(t, process.ErrNilCode, err)

	function, err := parser.GetFunction()
	assert.Equal(t, "", function)
	assert.Equal(t, process.ErrNilFunction, err)
}

func TestAtArgumentParser_GetArguments(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("aaaa@a@b@c")
	assert.Nil(t, err)

	args, err := parser.GetArguments()
	assert.Nil(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, 3, len(args))
}

func TestAtArgumentParser_GetArgumentsEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("aaaa")
	assert.Nil(t, err)

	args, err := parser.GetArguments()
	assert.Nil(t, err)
	assert.NotNil(t, args)
	assert.Equal(t, 0, len(args))
}

func TestAtArgumentParser_GetCode(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("bbbbbbbb@aaaa")
	assert.Nil(t, err)

	code, err := parser.GetCode()
	assert.Nil(t, err)
	assert.NotNil(t, code)
	assert.Equal(t, []byte("bbbbbbbb"), code)
}

func TestAtArgumentParser_GetCodeEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("@aaaa")
	assert.Equal(t, process.ErrStringSplitFailed, err)

	code, err := parser.GetCode()
	assert.Equal(t, process.ErrNilCode, err)
	assert.Nil(t, code)
}

func TestAtArgumentParser_GetFunction(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("bbbbbbbb@aaaa")
	assert.Nil(t, err)

	function, err := parser.GetFunction()
	assert.Nil(t, err)
	assert.NotNil(t, function)
	assert.Equal(t, []byte("bbbbbbbb"), []byte(function))
}

func TestAtArgumentParser_GetFunctionEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("@a")
	assert.Equal(t, process.ErrStringSplitFailed, err)

	function, err := parser.GetFunction()
	assert.Equal(t, process.ErrNilFunction, err)
	assert.Equal(t, 0, len(function))
}

func TestAtArgumentParser_ParseData(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("ab")
	assert.Nil(t, err)
}

func TestAtArgumentParser_ParseDataEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData("")
	assert.Equal(t, process.ErrStringSplitFailed, err)
}

func TestAtArgumentParser_CreateDataFromStorageUpdate(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	data := parser.CreateDataFromStorageUpdate(nil)
	assert.Equal(t, 0, len(data))

	test := []byte("aaaa")
	stUpd := vmcommon.StorageUpdate{Offset: test, Data: test}
	stUpdates := make([]*vmcommon.StorageUpdate, 0)
	stUpdates = append(stUpdates, &stUpd, &stUpd, &stUpd)
	result := ""
	sep := "@"
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)
	result = result + sep
	result = result + hex.EncodeToString(test)

	data = parser.CreateDataFromStorageUpdate(stUpdates)

	assert.Equal(t, result, data)
}

func TestAtArgumentParser_GetStorageUpdatesEmptyData(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	stUpdates, err := parser.GetStorageUpdates("")

	assert.Nil(t, stUpdates)
	assert.Equal(t, process.ErrStringSplitFailed, err)
}

func TestAtArgumentParser_GetStorageUpdatesWrongData(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	test := "test"
	result := ""
	sep := "@"
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test

	stUpdates, err := parser.GetStorageUpdates(result)

	assert.Nil(t, stUpdates)
	assert.Equal(t, process.ErrInvalidDataInput, err)
}

func TestAtArgumentParser_GetStorageUpdates(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	test := "aaaa"
	result := ""
	sep := "@"
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	result = result + sep
	result = result + test
	stUpdates, err := parser.GetStorageUpdates(result)

	assert.Nil(t, err)
	for i := 0; i < 2; i++ {
		assert.Equal(t, test, hex.EncodeToString(stUpdates[i].Data))
		assert.Equal(t, test, hex.EncodeToString(stUpdates[i].Offset))
	}
}
