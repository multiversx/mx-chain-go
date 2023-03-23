package factory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
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
		HttpClient: httpClient,
		Marshaller: args.Marshaller,
	}

	return notifier.NewEventNotifier(notifierArgs)
}

func checkInputArgs(args *EventNotifierFactoryArgs) error {
	if check.IfNil(args.Marshaller) {
		return core.ErrNilMarshalizer
	}

	return nil
}
