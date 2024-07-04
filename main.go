package main

import (
	"github.com/shutter-network/nethermind-tests/rpc"
	"time"
)

func main() {
	tick := time.NewTicker(5 * time.Second)

	for range tick.C {
		println("Sending transaction")
		rpc.SendLegacyTx("https://erpc.chiado.staging.shutter.network")
		// todo add gnosis rpc
		// rpc.SendLegacyTx("GNOSIS_RPC_URL")
	}

}
