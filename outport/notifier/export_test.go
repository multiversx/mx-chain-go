package notifier

import nodeData "github.com/ElrondNetwork/elrond-go-core/data"

func (en *eventNotifier) GetLogEventsFromTransactionsPool(logs []*nodeData.LogData) []Event {
	return en.getLogEventsFromTransactionsPool(logs)
}
