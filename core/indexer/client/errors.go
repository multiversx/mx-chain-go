package client

import "errors"

// ErrCouldNotCreatePolicy -
var ErrCouldNotCreatePolicy = errors.New("could not create policy")

// ErrNoElasticUrlProvided signals that was not provided a url to the elasticsearch database
var ErrNoElasticUrlProvided = errors.New("no elastic url provided")

// ErrBackOff signals that an error was received from the server
var ErrBackOff = errors.New("back off something is not working well")
