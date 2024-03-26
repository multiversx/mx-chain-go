package sovereign

import (
	"github.com/multiversx/mx-chain-go/process/block/sovereign"
)

// DataDecoderCreator is an interface for creating data decoder
type DataDecoderCreator interface {
	CreateDataCodec() sovereign.DataDecoderHandler
	IsInterfaceNil() bool
}

// TopicsCheckerCreator is an interface for creating topics checker
type TopicsCheckerCreator interface {
	CreateTopicsChecker() sovereign.TopicsCheckerHandler
	IsInterfaceNil() bool
}
