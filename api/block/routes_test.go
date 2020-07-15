package block

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToUint64(t *testing.T) {
	t.Parallel()

	intValue, err := strconv.ParseUint("", 10, 64)

	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), intValue)
}
