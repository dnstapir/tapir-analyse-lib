package common

import "errors"

var ErrFatal = errors.New("fatal")
var ErrBadHandle = errors.New("bad handle")
var ErrBadParam = errors.New("bad parameter")
var ErrBadFlag = errors.New("bad observation flag")
var ErrBadKey = errors.New("bad nats key")
var ErrBadJWK = errors.New("bad JWK key")
var ErrBadJSON = errors.New("bad json")
var ErrNotCompleted = errors.New("could not complete operation")
