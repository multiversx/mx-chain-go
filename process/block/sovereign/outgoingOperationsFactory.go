package sovereign

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-go/config"
)

// CreateOutgoingOperationsFormatter creates an outgoing operations formatter
func CreateOutgoingOperationsFormatter(
	events []config.SubscribedEvent,
	pubKeyConverter core.PubkeyConverter,
	dataCodec DataCodecProcessor,
) (OutgoingOperationsFormatter, error) {
	subscribedEvents, err := getSubscribedEvents(events, pubKeyConverter)
	if err != nil {
		return nil, err
	}

	return NewOutgoingOperationsFormatter(subscribedEvents, dataCodec)
}

func getSubscribedEvents(events []config.SubscribedEvent, pubKeyConverter core.PubkeyConverter) ([]SubscribedEvent, error) {
	ret := make([]SubscribedEvent, len(events))
	for idx, event := range events {
		addressesMap, err := getAddressesMap(event.Addresses, pubKeyConverter)
		if err != nil {
			return nil, fmt.Errorf("%w for event at index = %d", err, idx)
		}

		ret[idx] = SubscribedEvent{
			Identifier: []byte(event.Identifier),
			Addresses:  addressesMap,
		}
	}

	return ret, nil
}

func getAddressesMap(addresses []string, pubKeyConverter core.PubkeyConverter) (map[string]string, error) {
	numAddresses := len(addresses)
	if numAddresses == 0 {
		return nil, errNoSubscribedAddresses
	}

	addressesMap := make(map[string]string, numAddresses)
	for _, encodedAddr := range addresses {
		decodedAddr, errDecode := pubKeyConverter.Decode(encodedAddr)
		if errDecode != nil {
			return nil, errDecode
		}

		addressesMap[string(decodedAddr)] = encodedAddr
	}

	if len(addressesMap) != numAddresses {
		return nil, errDuplicateSubscribedAddresses
	}

	return addressesMap, nil
}
