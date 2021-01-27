package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBenchmarksList(t *testing.T) {
	list := CreateBenchmarksList("../testdata")

	assert.Equal(t, 15, len(list))
}
