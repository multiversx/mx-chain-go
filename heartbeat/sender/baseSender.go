package sender

import (
	"fmt"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/heartbeat"
)

const minTimeBetweenSends = time.Second

type ArgBaseSender struct {
	Messenger                 heartbeat.P2PMessenger
	Marshaller                marshal.Marshalizer
	Topic                     string
	TimeBetweenSends          time.Duration
	TimeBetweenSendsWhenError time.Duration
}

type baseSender struct {
	timerHandler
	messenger                 heartbeat.P2PMessenger
	marshaller                marshal.Marshalizer
	topic                     string
	timeBetweenSends          time.Duration
	timeBetweenSendsWhenError time.Duration
}

func createBaseSender(args ArgBaseSender) baseSender {
	return baseSender{
		timerHandler: &timerWrapper{
			timer: time.NewTimer(args.TimeBetweenSends),
		},
		messenger:                 args.Messenger,
		marshaller:                args.Marshaller,
		topic:                     args.Topic,
		timeBetweenSends:          args.TimeBetweenSends,
		timeBetweenSendsWhenError: args.TimeBetweenSendsWhenError,
	}
}

func checkBaseSenderArgs(args ArgBaseSender) error {
	if check.IfNil(args.Messenger) {
		return heartbeat.ErrNilMessenger
	}
	if check.IfNil(args.Marshaller) {
		return heartbeat.ErrNilMarshaller
	}
	if len(args.Topic) == 0 {
		return heartbeat.ErrEmptySendTopic
	}
	if args.TimeBetweenSends < minTimeBetweenSends {
		return fmt.Errorf("%w for TimeBetweenSends", heartbeat.ErrInvalidTimeDuration)
	}
	if args.TimeBetweenSendsWhenError < minTimeBetweenSends {
		return fmt.Errorf("%w for TimeBetweenSendsWhenError", heartbeat.ErrInvalidTimeDuration)
	}

	return nil
}
