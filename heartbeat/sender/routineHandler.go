package sender

import (
	"context"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("heartbeat/sender")

type routineHandler struct {
	peerAuthenticationSender senderHandler
	heartbeatSender          senderHandler
	hardforkSender           hardforkHandler
	cancel                   func()
}

func newRoutineHandler(peerAuthenticationSender senderHandler, heartbeatSender senderHandler, hardforkSender hardforkHandler) *routineHandler {
	handler := &routineHandler{
		peerAuthenticationSender: peerAuthenticationSender,
		heartbeatSender:          heartbeatSender,
		hardforkSender:           hardforkSender,
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
		case <-ctx.Done():
			return
		}
	}
}

func (handler *routineHandler) closeProcessLoop() {
	handler.cancel()
}
