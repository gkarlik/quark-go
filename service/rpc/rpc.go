package rpc

import (
	"carlos/quark"
)

// RPC represents Remote Procedure Call mechanism
type RPC interface {
	StartRPCService(s quark.RPCService)
}
