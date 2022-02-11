package sender

import (
	"context"

	logger "github.com/ElrondNetwork/elrond-go-logger"
)

var log = logger.GetOrCreate("heartbeat/sender")

type routineHandler struct {
	peerAuthenticationSender senderHandler
	heartbeatSender          senderHandler
	cancel                   func()
}

func newRoutingHandler(peerAuthenticationSender senderHandler, heartbeatSender senderHandler) *routineHandler {
	handler := &routineHandler{
		peerAuthenticationSender: peerAuthenticationSender,
		heartbeatSender:          heartbeatSender,
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
	}()

	handler.peerAuthenticationSender.Execute()
	handler.heartbeatSender.Execute()

	for {
		select {
		case <-handler.peerAuthenticationSender.ShouldExecute():
			handler.peerAuthenticationSender.Execute()
		case <-handler.heartbeatSender.ShouldExecute():
			handler.heartbeatSender.Execute()
		case <-ctx.Done():
			return
		}
	}
}

func (handler *routineHandler) closeProcessLoop() {
	handler.cancel()
}
