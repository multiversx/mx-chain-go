package client

import "errors"

// ErrCouldNotCreatePolicy signals that the index policy hasn't been created
var ErrCouldNotCreatePolicy = errors.New("could not create policy")

// ErrNoElasticUrlProvided signals that the url to the elasticsearch database hasn't been provided
var ErrNoElasticUrlProvided = errors.New("no elastic url provided")

// ErrBackOff signals that an error was received from the server
var ErrBackOff = errors.New("back off something is not working well")
