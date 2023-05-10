package factory

import (
	"github.com/multiversx/mx-chain-communication-go/websocket/data"
	"github.com/multiversx/mx-chain-communication-go/websocket/factory"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/config"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/host"
	logger "github.com/multiversx/mx-chain-logger-go"
)

type ArgsHostDriverFactory struct {
	HostConfig config.HostDriverConfig
	Marshaller marshal.Marshalizer
}

var log = logger.GetOrCreate("outport/factory/hostdriver")

func CreateHostDriver(args ArgsHostDriverFactory) (outport.Driver, error) {
	wsHost, err := factory.CreateWebSocketHost(factory.ArgsWebSocketHost{
		WebSocketConfig: data.WebSocketConfig{
			URL:                args.HostConfig.URL,
			WithAcknowledge:    args.HostConfig.WithAcknowledge,
			IsServer:           args.HostConfig.IsServer,
			RetryDurationInSec: args.HostConfig.RetryDurationInSec,
			BlockingAckOnError: args.HostConfig.BlockingAckOnError,
		},
		Marshaller: args.Marshaller,
		Log:        log,
	})
	if err != nil {
		return nil, err
	}

	return host.NewHostDriver(host.ArgsHostDriver{
		Marshaller: args.Marshaller,
		SenderHost: wsHost,
		Log:        log,
	})
}
