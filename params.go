// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"github.com/fonero-project/fnod/chaincfg"
)

// activeNetParams is a pointer to the parameters specific to the
// currently active Fonero network.
var activeNetParams = &mainNetParams

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*chaincfg.Params
	rpcPort string
}

// mainNetParams contains parameters specific to the main network
// (wire.MainNet).  NOTE: The RPC port is intentionally different than the
// reference implementation because fnod does not handle wallet requests.  The
// separate wallet process listens on the well-known port and forwards requests
// it does not handle on to fnod.  This approach allows the wallet process
// to emulate the full reference implementation RPC API.
var mainNetParams = params{
	Params:  &chaincfg.MainNetParams,
	rpcPort: "9209",
}

// testNetParams contains parameters specific to the test network (version )
// (wire.TestNet).
var testNetParams = params{
	Params:  &chaincfg.TestNetParams,
	rpcPort: "19209",
}

// simNetParams contains parameters specific to the simulation test network
// (wire.SimNet).
var simNetParams = params{
	Params:  &chaincfg.SimNetParams,
	rpcPort: "19556",
}

// regNetParams contains parameters specific to the regression test
// network (wire.RegNet).
var regNetParams = params{
	Params:  &chaincfg.RegNetParams,
	rpcPort: "18656",
}
