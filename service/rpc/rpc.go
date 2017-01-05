package rpc

import (
	"github.com/gkarlik/quark"
)

// RPC represents Remote Procedure Call mechanism
type RPC interface {
	StartRPCService(s quark.RPCService)
}
