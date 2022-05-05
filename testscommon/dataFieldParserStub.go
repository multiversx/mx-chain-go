package testscommon

import datafield "github.com/ElrondNetwork/elrond-vm-common/parsers/dataField"

type DataFieldParserStub struct {
	ParseCalled func(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData
}

func (df *DataFieldParserStub) Parse(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData {
	if df.ParseCalled != nil {
		return df.ParseCalled(dataField, sender, receiver)
	}

	return nil
}
