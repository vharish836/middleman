package mcservice

import (
	"errors"
)

var errNumParameter = errors.New("not enough parameters to process, refer to help")
var errParameter = errors.New("invalid parameter value")
var errInternal = errors.New("internal server error")
