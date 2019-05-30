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
	assert.Nil(t, function)
	assert.Equal(t, process.ErrNilFunction, err)
}

func TestAtArgumentParser_GetArguments(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_GetArgumentsNil(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_GetCode(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_GetCodeNil(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_GetFunction(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_GetFunctionNil(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_ParseData(t *testing.T) {
	t.Parallel()
}

func TestAtArgumentParser_ParseDataNil(t *testing.T) {
	t.Parallel()
}
