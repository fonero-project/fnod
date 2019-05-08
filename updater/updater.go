// Copyright (c) 2019 The Fonero developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package updater

import (
	"fmt"
	"path/filepath"

	"github.com/frankbraun/codechain/hashchain"
	"github.com/fonero-project/fnod/wire"
)

// Head returns the current hashchain head for the given net.
func Head(net wire.CurrencyNet) (*[32]byte, error) {
	var dir string
	switch net {
	case wire.MainNet:
		dir = ".codechain_mainnet"
	case wire.TestNet:
		dir = ".codechain_testnet"
	default:
		return nil, fmt.Errorf("wire.CurrencyNet %v not supported", net)
	}
	filename := filepath.Join(dir, "hashchain")
	hc, err := hashchain.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	defer hc.Close()
	h := hc.Head()
	return &h, nil
}
