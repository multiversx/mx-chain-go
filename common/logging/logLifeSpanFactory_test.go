package logging

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
)

func TestNewTypeLogLifeSpanFactory_CreateLogLifeSpannerShouldWork(t *testing.T) {
	factory := &typeLogLifeSpanFactory{}

	args := LogLifeSpanFactoryArgs{
		EpochStartNotifierWithConfirm: &testscommon.EpochStartNotifierStub{},
		RecreateEvery:                 10,
	}

	args.LifeSpanType = "second"
	lls, err := factory.CreateLogLifeSpanner(args)
	assert.Nil(t, err)
	assert.NotNil(t, lls)
	sls, ok := lls.(*secondsLifeSpanner)
	assert.NotNil(t, sls)
	assert.True(t, ok)

	args.LifeSpanType = "epoch"
	lls, err = factory.CreateLogLifeSpanner(args)
	assert.Nil(t, err)
	assert.NotNil(t, lls)
	els, ok := lls.(*epochsLifeSpanner)
	assert.NotNil(t, els)
	assert.True(t, ok)

	args.LifeSpanType = "megabyte"
	lls, err = factory.CreateLogLifeSpanner(args)
	assert.Nil(t, err)
	assert.NotNil(t, lls)
	mls, ok := lls.(*sizeLifeSpanner)
	assert.NotNil(t, mls)
	assert.True(t, ok)
}

func TestNewTypeLogLifeSpanFactory_CreateLogLifeSpannerInvalidShouldFail(t *testing.T) {
	factory := &typeLogLifeSpanFactory{}

	args := LogLifeSpanFactoryArgs{
		EpochStartNotifierWithConfirm: &testscommon.EpochStartNotifierStub{},
		LifeSpanType:                  "invalid",
		RecreateEvery:                 10,
	}

	lls, err := factory.CreateLogLifeSpanner(args)
	assert.Nil(t, lls)
	assert.Equal(t, ErrUnsupportedLogLifeSpanType, err)
}
