package factory

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/enableEpochsHandlerMock"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/vm"
	"github.com/multiversx/mx-chain-go/vm/mock"
	"github.com/multiversx/mx-chain-go/vm/systemSmartContracts/defaults"
	wasmConfig "github.com/multiversx/mx-chain-vm-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockNewSystemScFactoryArgs() ArgsNewSystemSCFactory {
	gasMap := wasmConfig.MakeGasMapForTests()
	gasMap = defaults.FillGasMapInternal(gasMap, 1)
	gasSchedule := testscommon.NewGasScheduleNotifierMock(gasMap)
	return ArgsNewSystemSCFactory{
		SystemEI:            &mock.SystemEIStub{},
		Economics:           &mock.EconomicsHandlerStub{},
		SigVerifier:         &mock.MessageSignVerifierMock{},
		GasSchedule:         gasSchedule,
		NodesConfigProvider: &mock.NodesConfigProviderStub{},
		Marshalizer:         &mock.MarshalizerMock{},
		Hasher:              &hashingMocks.HasherMock{},
		SystemSCConfig: &config.SystemSmartContractsConfig{
			ESDTSystemSCConfig: config.ESDTSystemSCConfig{
				BaseIssuingCost: "100000000",
				OwnerAddress:    "aaaaaa",
			},
			GovernanceSystemSCConfig: config.GovernanceSystemSCConfig{
				V1: config.GovernanceSystemSCConfigV1{
					NumNodes:         3,
					MinPassThreshold: 1,
					MinQuorum:        2,
					MinVetoThreshold: 2,
					ProposalCost:     "100",
				},
				Active: config.GovernanceSystemSCConfigActive{
					ProposalCost:     "500",
					MinQuorum:        0.5,
					MinPassThreshold: 0.5,
					MinVetoThreshold: 0.5,
					LostProposalFee:  "1",
				},
				OwnerAddress:                 "3132333435363738393031323334353637383930313233343536373839303234",
				MaxVotingDelayPeriodInEpochs: 30,
			},
			StakingSystemSCConfig: config.StakingSystemSCConfig{
				GenesisNodePrice:                     "1000",
				UnJailValue:                          "10",
				MinStepValue:                         "10",
				MinStakeValue:                        "1",
				UnBondPeriod:                         1,
				NumRoundsWithoutBleed:                1,
				MaximumPercentageToBleed:             1,
				BleedPercentagePerRound:              1,
				MaxNumberOfNodesForStake:             100,
				ActivateBLSPubKeyMessageVerification: false,
				MinUnstakeTokensValue:                "1",
			},
			DelegationSystemSCConfig: config.DelegationSystemSCConfig{
				MinServiceFee: 0,
				MaxServiceFee: 10000,
			},
			DelegationManagerSystemSCConfig: config.DelegationManagerSystemSCConfig{
				MinCreationDeposit:  "10",
				MinStakeAmount:      "10",
				ConfigChangeAddress: "3132333435363738393031323334353637383930313233343536373839303234",
			},
		},
		AddressPubKeyConverter: &testscommon.PubkeyConverterMock{},
		ShardCoordinator:       &mock.ShardCoordinatorStub{},
		EnableEpochsHandler:    &enableEpochsHandlerMock.EnableEpochsHandlerStub{},
	}
}

func TestNewSystemSCFactory_NilSystemEI(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.SystemEI = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilSystemEnvironmentInterface))
}

func TestNewSystemSCFactory_NilSigVerifier(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.SigVerifier = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilMessageSignVerifier))
}

func TestNewSystemSCFactory_NilNodesConfigProvider(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.NodesConfigProvider = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilNodesConfigProvider))
}

func TestNewSystemSCFactory_NilMarshalizer(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.Marshalizer = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilMarshalizer))
}

func TestNewSystemSCFactory_NilHasher(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.Hasher = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilHasher))
}

func TestNewSystemSCFactory_NilEconomicsData(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.Economics = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilEconomicsData))
}

func TestNewSystemSCFactory_NilSystemScConfig(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.SystemSCConfig = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilSystemSCConfig))
}

func TestNewSystemSCFactory_NilPubKeyConverter(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.AddressPubKeyConverter = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, scFactory)
	assert.True(t, errors.Is(err, vm.ErrNilAddressPubKeyConverter))
}

func TestNewSystemSCFactory_NilShardCoordinator(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.ShardCoordinator = nil
	scFactory, err := NewSystemSCFactory(arguments)

	assert.True(t, check.IfNil(scFactory))
	assert.True(t, errors.Is(err, vm.ErrNilShardCoordinator))
}

func TestNewSystemSCFactory_Ok(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, err := NewSystemSCFactory(arguments)

	assert.Nil(t, err)
	assert.NotNil(t, scFactory)
}

func TestNewSystemSCFactory_GasScheduleChangeMissingElementsShouldNotPanic(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)

	gasSchedule, err := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	delete(gasSchedule["MetaChainSystemSCsCost"], "UnstakeTokens")
	require.Nil(t, err)

	scFactory.GasScheduleChange(gasSchedule)

	assert.Equal(t, uint64(1), scFactory.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
}

func TestNewSystemSCFactory_GasScheduleChangeShouldWork(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, fmt.Sprintf("should have not panicked: %v", r))
		}
	}()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)

	gasSchedule, err := common.LoadGasScheduleConfig("../../cmd/node/config/gasSchedules/gasScheduleV3.toml")
	require.Nil(t, err)

	scFactory.GasScheduleChange(gasSchedule)

	assert.Equal(t, uint64(5000000), scFactory.gasCost.MetaChainSystemSCsCost.UnStakeTokens)
}

func TestSystemSCFactory_CreateWithBadDelegationManagerConfigChangeAddressShouldError(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	arguments.SystemSCConfig.DelegationManagerSystemSCConfig.ConfigChangeAddress = "not a hex string"
	scFactory, _ := NewSystemSCFactory(arguments)

	container, err := scFactory.Create()

	assert.True(t, check.IfNil(container))
	assert.True(t, errors.Is(err, vm.ErrInvalidAddress))
}

func TestSystemSCFactory_Create(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)

	container, err := scFactory.Create()
	assert.Nil(t, err)
	require.NotNil(t, container)
	assert.Equal(t, 6, container.Len())
}

func TestSystemSCFactory_CreateForGenesis(t *testing.T) {
	t.Parallel()

	arguments := createMockNewSystemScFactoryArgs()
	scFactory, _ := NewSystemSCFactory(arguments)

	container, err := scFactory.CreateForGenesis()
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
