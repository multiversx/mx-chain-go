package sovereign

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-go/config"
)

const (
	addressLen = 32
)

func CrateOutgoingOperationsFormatter(events []config.SubscribedEvent) (OutgoingOperationsFormatter, error) {
	subscribedEvents, err := getSubscribedEvents(events)
	if err != nil {
		return nil, err
	}

	return NewOutgoingOperationsCreator(subscribedEvents)
}

func getSubscribedEvents(events []config.SubscribedEvent) ([]SubscribedEvent, error) {
	ret := make([]SubscribedEvent, len(events))
	for idx, event := range events {
		addressesMap, err := getAddressesMap(event.Addresses)
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

func getAddressesMap(addresses []string) (map[string]string, error) {
	numAddresses := len(addresses)
	if numAddresses == 0 {
		return nil, errNoSubscribedAddresses
	}

	pubKeyConv, err := pubkeyConverter.NewBech32PubkeyConverter(addressLen, core.DefaultAddressPrefix)
	if err != nil {
		return nil, err
	}

	addressesMap := make(map[string]string, numAddresses)
	for _, encodedAddr := range addresses {
		decodedAddr, errDecode := pubKeyConv.Decode(encodedAddr)
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
