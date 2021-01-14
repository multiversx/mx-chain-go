package hostParameters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHostParametersGetter_GetParameterStringTable(t *testing.T) {
	testVersion := "assessment-go1.15.5/linux-amd64/1122ccdd00"
	hpg := NewHostParameterGetter(testVersion)

	str := hpg.GetParameterStringTable()
	fmt.Printf(str)
	assert.Contains(t, str, versionMarker)
	assert.Contains(t, str, testVersion)
	assert.Contains(t, str, modelMarker)
	assert.Contains(t, str, numLogicalMarker)
	assert.Contains(t, str, maxFreqMarker)
	assert.Contains(t, str, freqMarker)
	assert.Contains(t, str, flagsMarker)
	assert.Contains(t, str, sizeMarker)
}
