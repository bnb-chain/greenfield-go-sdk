package types

import (
	"runtime"
)

const (
	Denomm         = "BNB"
	libraryName    = "greenfield-go-sdk"
	libraryVersion = "v0.0.1"

	libraryUserAgentPrefix = "Greenfield (" + runtime.GOOS + "; " + runtime.GOARCH + ") "
	libraryUserAgent       = libraryUserAgentPrefix + libraryName + "/" + libraryVersion
)
