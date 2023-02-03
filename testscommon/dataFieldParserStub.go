package testscommon

import datafield "github.com/multiversx/mx-chain-vm-common-go/parsers/dataField"

// DataFieldParserStub -
type DataFieldParserStub struct {
	ParseCalled func(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData
}

// Parse -
func (df *DataFieldParserStub) Parse(dataField []byte, sender, receiver []byte, numOfShards uint32) *datafield.ResponseParseData {
	if df.ParseCalled != nil {
		return df.ParseCalled(dataField, sender, receiver, numOfShards)
	}

	return nil
}
