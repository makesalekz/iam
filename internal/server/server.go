package server

import (
	"github.com/google/wire"
)

// ProviderSet is server providers.
//
//nolint:gochecknoglobals // this global variable is required for wire
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer, NewCronServer)
