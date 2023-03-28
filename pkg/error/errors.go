package error

import "errors"

var (
	ErrorKeyManagerNotInit = errors.New("Key manager is not initialized yet ")
	ErrorUrlNotProvided    = errors.New("Url address not provided yet ")
	ErrorUrlsMismatch      = errors.New("Number of RPC and GRPC Urls does not match ")
)
