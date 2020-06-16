package factory

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/vm"
	"github.com/ElrondNetwork/elrond-go/vm/mock"
	"github.com/ElrondNetwork/elrond-go/vm/systemSmartContracts/defaults"
	"github.com/stretchr/testify/assert"
)

func createMockNewSystemScFactoryArgs() ArgsNewSystemSCFactory {
	gasSchedule := make(map[string]map[string]uint64)
	gasSchedule = defaults.FillGasMapInternal(gasSchedule, 1)

	return ArgsNewSystemSCFactory{
		SystemEI:            &mock.SystemEIStub{},
		ValidatorSettings:   &mock.ValidatorSettingsStub{},
		SigVerifier:         &mock.MessageSignVerifierMock{},
		GasMap:              gasSchedule,
		NodesConfigProvider: &mock.NodesConfigProviderStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &mock.HasherMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "100000000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				ProposalCost:     "500",
				NumNodes:         100,
				MinQuorum:        50,
				MinPassThreshold: 50,
				MinVetoThreshold: 50,
			},
		},
	}
}

func TestNewSystemSCFactory_NilSystemEI(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.SystemEI = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.Equal(t, vm.ErrNilSystemEnvironmentInterface, err)
}

func TestNewSystemSCFactory_NilEconomicsData(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.ValidatorSettings = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.Equal(t, vm.ErrNilEconomicsData, err)
}

func TestNewSystemSCFactory_Ok(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, err)
	assert.NotNil(t, scFactory)
}

func TestSystemSCFactory_Create(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)

	container, err := scFactory.Create()
	assert.Nil(t, err)
	assert.Equal(t, 4, container.Len())
}

func TestSystemSCFactory_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)
	assert.False(t, scFactory.IsInterfaceNil())

	scFactory = nil
	assert.True(t, check.IfNil(scFactory))
}
