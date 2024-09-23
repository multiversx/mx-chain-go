package hostParameters

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostParametersGetter_GetParameterStringTable(t *testing.T) {
	testVersion := "assessment-go1.15.5/linux-amd64/1122ccdd00"
	hpg := NewHostParameterGetter(testVersion)

	hi := hpg.GetHostInfo()
	assert.NotNil(t, hi)
	assert.Equal(t, testVersion, hi.AppVersion)
}
