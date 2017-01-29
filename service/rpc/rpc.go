package rpc

import (
	"github.com/gkarlik/quark-go"
)

// RPC represents Remote Procedure Call mechanism
type RPC interface {
	Start(s quark.RPCService)
	Stop()
}
