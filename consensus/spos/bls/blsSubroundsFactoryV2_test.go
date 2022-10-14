package bls_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus/mock"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
	"github.com/stretchr/testify/assert"
)

func TestNewSubroundsFactoryV2_ShouldErrNilFactory(t *testing.T) {
	t.Parallel()

	fct, err := bls.NewSubroundsFactoryV2(nil)
	assert.Nil(t, fct)
	assert.Equal(t, spos.ErrNilFactory, err)
}

func TestNewSubroundsFactoryV2_ShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	fct := initFactoryWithContainer(container)

	fctV2, err := bls.NewSubroundsFactoryV2(fct)
	assert.NotNil(t, fctV2)
	assert.Nil(t, err)
}

func TestFactoryV2_GenerateSubroundBlockShouldFailWhenNewSubroundFail(t *testing.T) {
	t.Parallel()

	fct := *initFactory()
	fct.Worker().(*mock.SposWorkerMock).GetConsensusStateChangedChannelsCalled = func() chan bool {
		return nil
	}

	fctV2, _ := bls.NewSubroundsFactoryV2(&fct)
	err := fctV2.GenerateBlockSubround()

	assert.NotNil(t, err)
}

func TestFactoryV2_GenerateSubroundBlockShouldWork(t *testing.T) {
	t.Parallel()

	container := mock.InitConsensusCore()
	fct := *initFactoryWithContainer(container)

	fctV2, _ := bls.NewSubroundsFactoryV2(&fct)
	err := fctV2.GenerateBlockSubround()

	assert.Nil(t, err)
}
