package factory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/outport"
	"github.com/ElrondNetwork/elrond-go/outport/notifier"
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
