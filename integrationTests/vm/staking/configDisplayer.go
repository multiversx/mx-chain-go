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

func getAllPubKeys(validatorsMap map[uint32][][]byte) [][]byte {
	allValidators := make([][]byte, 0)
	for _, validatorsInShard := range validatorsMap {
		allValidators = append(allValidators, validatorsInShard...)
	}

	return allValidators
}

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
	lines = append(lines, display.NewLineData(true, []string{"eligible", fmt.Sprintf("Total: %d", len(getAllPubKeys(config.eligible))), "All shards"}))
	lines = append(lines, display.NewLineData(true, []string{"waiting", fmt.Sprintf("Total: %d", len(getAllPubKeys(config.waiting))), "All shards"}))
	lines = append(lines, display.NewLineData(true, []string{"leaving", fmt.Sprintf("Total: %d", len(getAllPubKeys(config.leaving))), "All shards"}))
	lines = append(lines, display.NewLineData(true, []string{"shuffled", fmt.Sprintf("Total: %d", len(getAllPubKeys(config.shuffledOut))), "All shards"}))

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
		horizontalLineAfter := idx == len(pubKeysToDisplay)-1
		line := display.NewLineData(horizontalLineAfter, []string{list, string(pk), strconv.Itoa(int(shardID))})
		lines = append(lines, line)
	}
	lines = append(lines, display.NewLineData(true, []string{list, fmt.Sprintf("Total: %d", len(pubKeys)), strconv.Itoa(int(shardID))}))

	return lines
}

func displayValidators(list string, pubKeys [][]byte) {
	pubKeysToDisplay := getShortPubKeysList(pubKeys)

	lines := make([]*display.LineData, 0)
	tableHeader := []string{"List", "Pub key"}
	for idx, pk := range pubKeysToDisplay {
		horizontalLineAfter := idx == len(pubKeysToDisplay)-1
		lines = append(lines, display.NewLineData(horizontalLineAfter, []string{list, string(pk)}))
	}
	lines = append(lines, display.NewLineData(true, []string{list, fmt.Sprintf("Total: %d", len(pubKeys))}))

	headline := display.Headline(fmt.Sprintf("%s list", list), "", delimiter)
	table, _ := display.CreateTableString(tableHeader, lines)
	fmt.Println(fmt.Sprintf("%s \n%s", headline, table))
}
