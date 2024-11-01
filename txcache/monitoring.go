package txcache

import (
	"fmt"
	"strings"

	logger "github.com/multiversx/mx-chain-logger-go"
)

func displaySendersScoreHistogram(scoreGroups [][]*txListForSender) {
	if log.GetLevel() > logger.LogDebug {
		return
	}

	stringBuilder := strings.Builder{}

	for i, group := range scoreGroups {
		if len(group) == 0 {
			continue
		}

		stringBuilder.WriteString(fmt.Sprintf("#%d: %d; ", i, len(group)))
	}

	log.Debug("displaySendersScoreHistogram()", "histogram", stringBuilder.String())
}

func displaySelectionOutcome(contextualLogger logger.Logger, selection []*WrappedTransaction) {
	if contextualLogger.GetLevel() > logger.LogTrace {
		return
	}

	if len(selection) > 0 {
		contextualLogger.Trace("displaySelectionOutcome() - transactions (as newline-separated JSON):")
		contextualLogger.Trace(marshalTransactionsToNewlineDelimitedJson(selection))
	} else {
		contextualLogger.Trace("displaySelectionOutcome() - transactions: none")
	}
}
