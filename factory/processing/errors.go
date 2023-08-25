package processing

import "errors"

var errGenesisMetaBlockDoesNotExist = errors.New("genesis meta block does not exist")

var errInvalidGenesisMetaBlock = errors.New("genesis meta block invalid, should be of type meta header handler")
