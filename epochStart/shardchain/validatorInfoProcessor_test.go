package shardchain

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
)

func createDefaultArguments() ArgValidatorInfoProcessor {
	defaultArgs := ArgValidatorInfoProcessor{
		MiniBlocksPool:               nil,
		Marshalizer:                  &mock.MarshalizerMock{},
		ValidatorStatisticsProcessor: nil,
		Requesthandler:               nil,
		Hasher:                       nil,
	}

	return defaultArgs
}

func TestNewValidatorInfoProcessor_NilValidatorStatisticsProcessorShouldErr(t *testing.T) {

}

func TestNewValidatorInfoProcessor_NilMarshalizerShouldErr(t *testing.T) {

}

func TestNewValidatorInfoProcessor_NilMiniBlocksPoolErr(t *testing.T) {

}

func TestNewValidatorInfoProcessor_NilRequesthandlerShouldErr(t *testing.T) {

}

func TestValidatorInfoProcessor_IsInterfaceNil(t *testing.T) {

}

func TestValidatorInfoProcessor_ProcessMetaBlock(t *testing.T) {

}
