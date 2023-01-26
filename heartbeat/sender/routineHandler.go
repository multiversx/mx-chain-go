package sender

import (
	"context"
	"time"

	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("heartbeat/sender")

type routineHandler struct {
	peerAuthenticationSender           senderHandler
	heartbeatSender                    senderHandler
	hardforkSender                     hardforkHandler
	delayAfterHardforkMessageBroadcast time.Duration
	cancel                             func()
}

func newRoutineHandler(peerAuthenticationSender senderHandler, heartbeatSender senderHandler, hardforkSender hardforkHandler) *routineHandler {
	handler := &routineHandler{
		peerAuthenticationSender:           peerAuthenticationSender,
		heartbeatSender:                    heartbeatSender,
		hardforkSender:                     hardforkSender,
		delayAfterHardforkMessageBroadcast: time.Minute,
	}

	var ctx context.Context
	ctx, handler.cancel = context.WithCancel(context.Background())
	go handler.processLoop(ctx)

	return handler
}

func (handler *routineHandler) processLoop(ctx context.Context) {
	defer func() {
		log.Debug("heartbeat's routine handler is closing...")

		handler.peerAuthenticationSender.Close()
		handler.heartbeatSender.Close()
		handler.hardforkSender.Close()
	}()

	handler.peerAuthenticationSender.Execute()
	handler.heartbeatSender.Execute()

	for {
		select {
		case <-handler.peerAuthenticationSender.ExecutionReadyChannel():
			handler.peerAuthenticationSender.Execute()
		case <-handler.heartbeatSender.ExecutionReadyChannel():
			handler.heartbeatSender.Execute()
		case <-handler.hardforkSender.ShouldTriggerHardfork():
			handler.hardforkSender.Execute()
			handler.waitAfterHarforkBroadcast(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (handler *routineHandler) waitAfterHarforkBroadcast(ctx context.Context) {
	timer := time.NewTimer(handler.delayAfterHardforkMessageBroadcast)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (handler *routineHandler) closeProcessLoop() {
	handler.cancel()
}
