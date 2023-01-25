package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/outport/notifier"
)

// EventNotifierFactoryArgs defines the args needed for event notifier creation
type EventNotifierFactoryArgs struct {
	Enabled           bool
	UseAuthorization  bool
	ProxyUrl          string
	Username          string
	Password          string
	RequestTimeoutSec int
	Marshaller        marshal.Marshalizer
	Hasher            hashing.Hasher
	PubKeyConverter   core.PubkeyConverter
}

// CreateEventNotifier will create a new event notifier client instance
func CreateEventNotifier(args *EventNotifierFactoryArgs) (outport.Driver, error) {
	if err := checkInputArgs(args); err != nil {
		return nil, err
	}

	httpClientArgs := notifier.HTTPClientWrapperArgs{
		UseAuthorization:  args.UseAuthorization,
		Username:          args.Username,
		Password:          args.Password,
		BaseUrl:           args.ProxyUrl,
		RequestTimeoutSec: args.RequestTimeoutSec,
	}
	httpClient, err := notifier.NewHTTPWrapperClient(httpClientArgs)
	if err != nil {
		return nil, err
	}

	notifierArgs := notifier.ArgsEventNotifier{
		HttpClient:      httpClient,
		Marshaller:      args.Marshaller,
		Hasher:          args.Hasher,
		PubKeyConverter: args.PubKeyConverter,
	}

	return notifier.NewEventNotifier(notifierArgs)
}

func checkInputArgs(args *EventNotifierFactoryArgs) error {
	if check.IfNil(args.Marshaller) {
		return core.ErrNilMarshalizer
	}
	if check.IfNil(args.Hasher) {
		return core.ErrNilHasher
	}
	if check.IfNil(args.PubKeyConverter) {
		return outport.ErrNilPubKeyConverter
	}

	return nil
}
