package esdtSupply

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/require"
)

func TestNewSuppliesProcessor(t *testing.T) {
	t.Parallel()

	_, err := NewSuppliesProcessor(nil, &testscommon.StorerStub{}, &testscommon.StorerStub{})
	require.Equal(t, core.ErrNilMarshalizer, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, nil, &testscommon.StorerStub{})
	require.Equal(t, core.ErrNilStore, err)

	_, err = NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &testscommon.StorerStub{}, nil)
	require.Equal(t, core.ErrNilStore, err)

	proc, err := NewSuppliesProcessor(&testscommon.MarshalizerMock{}, &testscommon.StorerStub{}, &testscommon.StorerStub{})
	require.Nil(t, err)
	require.NotNil(t, proc)
}
