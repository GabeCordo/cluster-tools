package common

import "errors"

var InternalError = errors.New("there was an internal error in the system")

var BadRequestType = errors.New("the request type does not match what was expected for this chan")

var BadResponseType = errors.New("the response type does not match what was expected for this chan")

var IllegalRequest = errors.New("the source is not permitted to send requests over this channel")

var UnknownRequest = errors.New("the request action is unknown to this thread")
