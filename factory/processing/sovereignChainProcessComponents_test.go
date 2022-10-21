package processing_test

import (
	"strings"
	"testing"

	errErd "github.com/ElrondNetwork/elrond-go/errors"
	processComp "github.com/ElrondNetwork/elrond-go/factory/processing"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	componentsMock "github.com/ElrondNetwork/elrond-go/testscommon/components"
	"github.com/stretchr/testify/assert"
)

func TestNewSovereignChainProcessComponentsFactory_ShouldErrNilProcessComponentsFactory(t *testing.T) {
	t.Parallel()

	scpcf, err := processComp.NewSovereignChainProcessComponentsFactory(nil)

	assert.Nil(t, scpcf)
	assert.Equal(t, errErd.ErrNilProcessComponentsFactory, err)
}

func TestNewSovereignChainProcessComponentsFactory_ShouldWork(t *testing.T) {
	t.Parallel()

	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	pcf, _ := processComp.NewProcessComponentsFactory(processArgs)

	scpcf, err := processComp.NewSovereignChainProcessComponentsFactory(pcf)

	assert.NotNil(t, scpcf)
	assert.Nil(t, err)
}

func TestSovereignChainProcessComponentsFactory_CreateWithInvalidTxAccumulatorTimeExpectError(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	shardCoordinator := testscommon.NewMultiShardsCoordinatorMock(2)
	processArgs := componentsMock.GetProcessComponentsFactoryArgs(shardCoordinator)
	processArgs.Config.Antiflood.TxAccumulator.MaxAllowedTimeInMilliseconds = 0

	pcf, _ := processComp.NewProcessComponentsFactory(processArgs)
	assert.NotNil(t, pcf)

	scpcf, _ := processComp.NewSovereignChainProcessComponentsFactory(pcf)
	assert.NotNil(t, scpcf)

	pc, err := scpcf.Create()

	assert.Nil(t, pc)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), process.ErrInvalidValue.Error()))
}
