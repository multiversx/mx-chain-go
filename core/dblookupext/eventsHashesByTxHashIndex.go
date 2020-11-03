//go:generate protoc -I=proto -I=$GOPATH/src -I=$GOPATH/src/github.com/ElrondNetwork/protobuf/protobuf  --gogoslick_out=. eventsHashesByTxHash.proto

package dblookupext

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/receipt"
	"github.com/ElrondNetwork/elrond-go/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
)

type eventsHashesByTxHash struct {
	marshalizer marshal.Marshalizer
	storer      storage.Storer
}

func newEventsHashesByTxHash(storer storage.Storer, marshalizer marshal.Marshalizer) *eventsHashesByTxHash {
	return &eventsHashesByTxHash{
		marshalizer: marshalizer,
		storer:      storer,
	}
}

func (eht *eventsHashesByTxHash) saveEventsHashes(epoch uint32, scrResults, receipts map[string]data.TransactionHandler) error {
	eventsHashes := eht.prepareEventsHashes(epoch, scrResults, receipts)
	for txHash, eventHashes := range eventsHashes {
		eventHashesBytes, err := eht.marshalizer.Marshal(eventHashes)
		if err != nil {
			continue
		}

		err = eht.storer.Put([]byte(txHash), eventHashesBytes)
		if err != nil {
			log.Debug("dblookupext.saveEventsHashes() cannot save eventsHashesByte",
				"error", err.Error())
			continue
		}
	}

	return nil
}

func (eht *eventsHashesByTxHash) getEventsHashesByTxHash(txHash []byte, epoch uint32) (*EventsHashesByTxHash, error) {
	rawBytes, err := eht.storer.GetFromEpoch(txHash, epoch)
	if err != nil {
		return nil, err
	}

	record := &EventsHashesByTxHash{}
	err = eht.marshalizer.Unmarshal(record, rawBytes)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (eht *eventsHashesByTxHash) prepareEventsHashes(epoch uint32, scrResults, receipts map[string]data.TransactionHandler) map[string]*EventsHashesByTxHash {
	eventsHashesMap := eht.prepareSmartContractResultsEvents(epoch, scrResults)

	for receiptHash, receiptHandler := range receipts {
		rec, ok := receiptHandler.(*receipt.Receipt)
		if !ok {
			log.Debug("dblookupext.prepareEventsHashes() cannot cast TransactionHandler to Receipt")
			continue
		}

		originalTxHash := string(rec.TxHash)
		if eventsHashes, ok := eventsHashesMap[originalTxHash]; ok {
			// append receipt hash in already create event
			eventsHashes.ReceiptsHash = []byte(receiptHash)
		} else {
			// create a new event for this hash
			eventsHashesMap[originalTxHash] = &EventsHashesByTxHash{
				ReceiptsHash: []byte(receiptHash),
			}
		}
	}

	return eventsHashesMap
}

func (eht *eventsHashesByTxHash) prepareSmartContractResultsEvents(
	epoch uint32,
	scrResults map[string]data.TransactionHandler,
) map[string]*EventsHashesByTxHash {
	eventsHashesMap := make(map[string]*EventsHashesByTxHash, 0)
	for scrHash, scrHandler := range scrResults {
		scrResult, ok := scrHandler.(*smartContractResult.SmartContractResult)
		if !ok {
			log.Debug("dblookupext.prepareEventsHashes() cannot cast TransactionHandler to SmartContractResult")
			continue
		}

		originalTxHash := string(scrResult.OriginalTxHash)

		if eventsHashes, ok := eventsHashesMap[originalTxHash]; ok {
			// append scr hash in already created event
			eventsHashes.ScrHashesEpoch[0].SmartContractResultsHashes = append(eventsHashes.ScrHashesEpoch[0].SmartContractResultsHashes, []byte(scrHash))
		} else {
			// create a new event for this hash
			eventsHashesMap[originalTxHash] = &EventsHashesByTxHash{
				ScrHashesEpoch: []*ScrHashesAndEpoch{
					{
						SmartContractResultsHashes: [][]byte{[]byte(scrHash)},
						Epoch:                      epoch,
					},
				},
			}
		}
	}

	for originalTxHash, events := range eventsHashesMap {
		rawBytes, err := eht.storer.Get([]byte(originalTxHash))
		if err != nil {
			continue
		}

		record := &EventsHashesByTxHash{}
		err = eht.marshalizer.Unmarshal(record, rawBytes)
		if err != nil {
			continue
		}

		// if a record exits in storage merge with current record
		events.ScrHashesEpoch = append(events.ScrHashesEpoch, record.ScrHashesEpoch...)
	}

	return eventsHashesMap
}
