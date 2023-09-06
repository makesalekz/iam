package biz

import (
	"iam/internal/data"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewDummy)


// Dummy .
type Dummy struct {
}

func NewDummy(d *data.Data) *Dummy {
	return &Dummy{}
}