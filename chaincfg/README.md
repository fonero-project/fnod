chaincfg
========

[![Build Status](http://img.shields.io/travis/fonero/fnod.svg)](https://travis-ci.org/fonero/fnod)
[![ISC License](http://img.shields.io/badge/license-ISC-blue.svg)](http://copyfree.org)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](http://godoc.org/github.com/fonero-project/fnod/chaincfg)

Package chaincfg defines chain configuration parameters for the three standard
Fonero networks and provides the ability for callers to define their own custom
Fonero networks.

Although this package was primarily written for fnod, it has intentionally been
designed so it can be used as a standalone package for any projects needing to
use parameters for the standard Fonero networks or for projects needing to
define their own network.

## Sample Use

```Go
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/fonero-project/fnod/fnoutil"
	"github.com/fonero-project/fnod/chaincfg"
)

var testnet = flag.Bool("testnet", false, "operate on the testnet Fonero network")

// By default (without -testnet), use mainnet.
var chainParams = &chaincfg.MainNetParams

func main() {
	flag.Parse()

	// Modify active network parameters if operating on testnet.
	if *testnet {
		chainParams = &chaincfg.TestNetParams
	}

	// later...

	// Create and print new payment address, specific to the active network.
	pubKeyHash := make([]byte, 20)
	addr, err := btcutil.NewAddressPubKeyHash(pubKeyHash, chainParams)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(addr)
}
```

## Installation and Updating

```bash
$ go get -u github.com/fonero-project/fnod/chaincfg
```

## License

Package chaincfg is licensed under the [copyfree](http://copyfree.org) ISC
License.
