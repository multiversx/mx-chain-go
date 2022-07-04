package logs

import "errors"

var errCannotCreateLogsFacade = errors.New("cannot create logs facade")
var errCannotLoadLogs = errors.New("cannot load log(s)")
var errCannotUnmarshalLog = errors.New("cannot unmarshal log")
