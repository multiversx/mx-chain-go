package guardian

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/epochNotifier"
	"github.com/stretchr/testify/require"
)

func TestNewAccountGuardianChecker(t *testing.T) {
	marshaller := &testscommon.MarshalizerMock{}
	en := &epochNotifier.EpochNotifierStub{}
	ga, err := NewGuardedAccount(marshaller, en)
	require.Nil(t, err)
	require.NotNil(t, ga)

	ga, err = NewGuardedAccount(nil, en)
	require.Equal(t, process.ErrNilMarshalizer, err)
	require.Nil(t, ga)

	ga, err = NewGuardedAccount(marshaller, nil)
	require.Equal(t, process.ErrNilEpochNotifier, err)
	require.Nil(t, ga)
}

func TestGuardedAccount_getActiveGuardian(t *testing.T) {

}

func TestGuardedAccount_getConfiguredGuardians(t *testing.T) {

}

func TestGuardedAccount_saveAccountGuardians(t *testing.T) {

}

func TestGuardedAccount_updateGuardians(t *testing.T) {

}

func TestGuardedAccount_setAccountGuardian(t *testing.T) {

}

func TestGuardedAccount_instantSetGuardian(t *testing.T) {

}

func TestGuardedAccount_EpochConfirmed(t *testing.T) {

}

func TestGuardedAccount_GetActiveGuardian(t *testing.T) {

}

func TestGuardedAccount_SetGuardian(t *testing.T) {

}

func TestAccountGuardianChecker_IsInterfaceNil(t *testing.T) {

}
