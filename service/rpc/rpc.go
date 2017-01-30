package rpc

import (
	"github.com/gkarlik/quark-go"
	"github.com/gkarlik/quark-go/system"
)

// RPC represents Remote Procedure Call server.
type RPC interface {
	Start(s quark.RPCService)
	Stop()

	system.Disposer
}
