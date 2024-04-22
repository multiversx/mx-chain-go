package metachain

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/display"
)

type tableDisplayer struct {
}

// NewTableDisplayer will create a component able to display tables in logger
func NewTableDisplayer() *tableDisplayer {
	return &tableDisplayer{}
}

// DisplayTable will display a table in the log
func (tb *tableDisplayer) DisplayTable(tableHeader []string, lines []*display.LineData, message string) {
	table, err := display.CreateTableString(tableHeader, lines)
	if err != nil {
		log.Error("could not create table", "tableHeader", tableHeader, "error", err)
		return
	}

	msg := fmt.Sprintf("%s\n%s", message, table)
	log.Debug(msg)
}

// IsInterfaceNil checks if the underlying pointer is nil
func (tb *tableDisplayer) IsInterfaceNil() bool {
	return tb == nil
}
