package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/stretchr/testify/assert"
	"testing"
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

	err = parser.ParseData([]byte("aaaa@a@b@c"))
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

	err = parser.ParseData([]byte("aaaa"))
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

	err = parser.ParseData([]byte("bbbbbbb@aaaa"))
	assert.Nil(t, err)

	code, err := parser.GetCode()
	assert.Nil(t, err)
	assert.NotNil(t, code)
	assert.Equal(t, []byte("bbbbbbb"), code)
}

func TestAtArgumentParser_GetCodeEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData([]byte("@aaaa"))
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

	err = parser.ParseData([]byte("bbbbbbb@aaaa"))
	assert.Nil(t, err)

	function, err := parser.GetFunction()
	assert.Nil(t, err)
	assert.NotNil(t, function)
	assert.Equal(t, "bbbbbbb", function)
}

func TestAtArgumentParser_GetFunctionEmpty(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData([]byte("@a"))
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

	err = parser.ParseData([]byte("a"))
	assert.Nil(t, err)
}

func TestAtArgumentParser_ParseDataNil(t *testing.T) {
	t.Parallel()

	parser, err := NewAtArgumentParser()
	assert.Nil(t, err)
	assert.NotNil(t, parser)

	err = parser.ParseData(nil)
	assert.Equal(t, process.ErrStringSplitFailed, err)
}
