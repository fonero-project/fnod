// Copyright (c) 2019 The Fonero developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

// addr2pkscript converts a Fonero address to a public key script.
package main

import (
	"fmt"
	"os"

	"github.com/fonero-project/fnod/txscript"
	"github.com/fonero-project/fnod/fnoutil"
)

func addr2PKScript(addrStr string) ([]byte, error) {
	addr, err := fnoutil.DecodeAddress(addrStr)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(addr)
	if err != nil {
		return nil, err
	}
	return pkScript, nil
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "%s: error: %s\n", os.Args[0], err)
	os.Exit(1)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s address\n", os.Args[0])
	os.Exit(2)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}
	pkScript, err := addr2PKScript(os.Args[1])
	if err != nil {
		fatal(err)
	}
	fmt.Printf("%x\n", pkScript)
}
