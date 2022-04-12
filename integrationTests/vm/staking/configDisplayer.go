package staking

import (
	"fmt"
	"strconv"

	"github.com/ElrondNetwork/elrond-go-core/display"
)

const (
	delimiter         = "#"
	maxPubKeysListLen = 6
)

// TODO: Make a subcomponent which will register to epoch notifier to display config only upon epoch change

func getShortPubKeysList(pubKeys [][]byte) [][]byte {
	pubKeysToDisplay := pubKeys
	if len(pubKeys) > maxPubKeysListLen {
		pubKeysToDisplay = make([][]byte, 0)
		pubKeysToDisplay = append(pubKeysToDisplay, pubKeys[:maxPubKeysListLen/2]...)
		pubKeysToDisplay = append(pubKeysToDisplay, [][]byte{[]byte("...")}...)
		pubKeysToDisplay = append(pubKeysToDisplay, pubKeys[len(pubKeys)-maxPubKeysListLen/2:]...)
	}

	return pubKeysToDisplay
}

func displayConfig(config nodesConfig) {
	lines := make([]*display.LineData, 0)

	for shard := range config.eligible {
		lines = append(lines, getDisplayableValidatorsInShard("eligible", config.eligible[shard], shard)...)
		lines = append(lines, getDisplayableValidatorsInShard("waiting", config.waiting[shard], shard)...)
		lines = append(lines, getDisplayableValidatorsInShard("leaving", config.leaving[shard], shard)...)
		lines = append(lines, getDisplayableValidatorsInShard("shuffled", config.shuffledOut[shard], shard)...)
		lines = append(lines, display.NewLineData(true, []string{}))
	}

	tableHeader := []string{"List", "Pub key", "Shard ID"}
	table, _ := display.CreateTableString(tableHeader, lines)
	headline := display.Headline("Nodes config", "", delimiter)
	fmt.Println(fmt.Sprintf("%s\n%s", headline, table))

	displayValidators("Auction", config.auction)
	displayValidators("Queue", config.queue)
}

func getDisplayableValidatorsInShard(list string, pubKeys [][]byte, shardID uint32) []*display.LineData {
	pubKeysToDisplay := getShortPubKeysList(pubKeys)

	lines := make([]*display.LineData, 0)
	for idx, pk := range pubKeysToDisplay {
		horizontalLine := idx == len(pubKeysToDisplay)-1
		line := display.NewLineData(horizontalLine, []string{list, string(pk), strconv.Itoa(int(shardID))})
		lines = append(lines, line)
	}

	return lines
}

func displayValidators(list string, pubKeys [][]byte) {
	pubKeysToDisplay := getShortPubKeysList(pubKeys)

	lines := make([]*display.LineData, 0)
	tableHeader := []string{"List", "Pub key"}
	for _, pk := range pubKeysToDisplay {
		lines = append(lines, display.NewLineData(false, []string{list, string(pk)}))
	}

	headline := display.Headline(fmt.Sprintf("%s list", list), "", delimiter)
	table, _ := display.CreateTableString(tableHeader, lines)
	fmt.Println(fmt.Sprintf("%s \n%s", headline, table))
}
