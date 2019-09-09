package interceptors

import (
	"github.com/ElrondNetwork/elrond-go/core/logger"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/ElrondNetwork/elrond-go/process"
)

var log = logger.DefaultLogger()

type messageChecker struct {
}

func (*messageChecker) checkMessage(message p2p.MessageP2P) error {
	if message == nil || message.IsInterfaceNil() {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	return nil
}
