// Copyright (c) 2013, 2014 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fnoutil_test

import (
	"testing"

	"github.com/fonero-project/fnod/chaincfg"
	"github.com/fonero-project/fnod/chaincfg/chainec"
	"github.com/fonero-project/fnod/fnoec"
	"github.com/fonero-project/fnod/fnoec/edwards"
	"github.com/fonero-project/fnod/fnoec/secp256k1"
	. "github.com/fonero-project/fnod/fnoutil"
)

func TestEncodeDecodeWIF(t *testing.T) {
	suites := []fnoec.SignatureType{
		fnoec.STEcdsaSecp256k1,
		fnoec.STEd25519,
		fnoec.STSchnorrSecp256k1,
	}
	for _, suite := range suites {
		var priv1, priv2 chainec.PrivateKey
		switch suite {
		case fnoec.STEcdsaSecp256k1:
			priv1, _ = secp256k1.PrivKeyFromBytes([]byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
				0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _ = secp256k1.PrivKeyFromBytes([]byte{
				0xdd, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
				0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
				0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
				0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		case fnoec.STEd25519:
			priv1, _, _ = edwards.PrivKeyFromScalar(edwards.Edwards(),
				[]byte{
					0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
					0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
					0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
					0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _, _ = edwards.PrivKeyFromScalar(edwards.Edwards(),
				[]byte{
					0x0c, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
					0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
					0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
					0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		case fnoec.STSchnorrSecp256k1:
			priv1, _ = secp256k1.PrivKeyFromBytes([]byte{
				0x0c, 0x28, 0xfc, 0xa3, 0x86, 0xc7, 0xa2, 0x27,
				0x60, 0x0b, 0x2f, 0xe5, 0x0b, 0x7c, 0xae, 0x11,
				0xec, 0x86, 0xd3, 0xbf, 0x1f, 0xbe, 0x47, 0x1b,
				0xe8, 0x98, 0x27, 0xe1, 0x9d, 0x72, 0xaa, 0x1d})

			priv2, _ = secp256k1.PrivKeyFromBytes([]byte{
				0xdd, 0xa3, 0x5a, 0x14, 0x88, 0xfb, 0x97, 0xb6,
				0xeb, 0x3f, 0xe6, 0xe9, 0xef, 0x2a, 0x25, 0x81,
				0x4e, 0x39, 0x6f, 0xb5, 0xdc, 0x29, 0x5f, 0xe9,
				0x94, 0xb9, 0x67, 0x89, 0xb2, 0x1a, 0x03, 0x98})
		}

		wif1, err := NewWIF(priv1, &chaincfg.MainNetParams, suite)
		if err != nil {
			t.Fatal(err)
		}
		wif2, err := NewWIF(priv2, &chaincfg.TestNetParams, suite)
		if err != nil {
			t.Fatal(err)
		}
		wif3, err := NewWIF(priv2, &chaincfg.SimNetParams, suite)
		if err != nil {
			t.Fatal(err)
		}
		wif4, err := NewWIF(priv2, &chaincfg.RegNetParams, suite)
		if err != nil {
			t.Fatal(err)
		}

		var tests []struct {
			wif     *WIF
			encoded string
		}

		switch suite {
		case fnoec.STEcdsaSecp256k1:
			tests = []struct {
				wif     *WIF
				encoded string
			}{
				{
					wif1,
					"PmQdMn8xafwaQouk8ngs1CccRCB1ZmsqQxBaxNR4vhQi5a5QB5716",
				},
				{
					wif2,
					"PtWVDUidYaiiNT5e2Sfb1Ah4evbaSopZJkkpFBuzkJYcYteugvdFg",
				},
				{
					wif3,
					"PsURoUb7FMeJQdTYea8pkbUQFBZAsxtfDcfTLGja5sCLZvLZWRtjK",
				},
				{
					wif4,
					"Pr9D8L8s9nG4AroRjbTGiRuYrweN1T8Dg9grAEeTEStZAPMnjxwCT",
				},
			}
		case fnoec.STEd25519:
			tests = []struct {
				wif     *WIF
				encoded string
			}{
				{
					wif1,
					"PmQfJXKC2ho1633ZiVbSdCZw1y68BVXYFpyE2UfDcbQN5xa3DByDn",
				},
				{
					wif2,
					"PtWVaBGeCfbFQfgqFew8YvdrSH5TH439K7rvpo3aWnSfDvyK8ijbK",
				},
				{
					wif3,
					"PsUSAB97uSWqSr4jsnQNJMRC2Y33iD7FDymZuss9rM6PExexSPyTQ",
				},
				{
					wif4,
					"Pr9DV2gsos8bD5QcxoipGBrLeJ8EqhLogWnxjqn2zvnbqRg9fZMHs",
				},
			}
		case fnoec.STSchnorrSecp256k1:
			tests = []struct {
				wif     *WIF
				encoded string
			}{
				{
					wif1,
					"PmQhFGVRUjeRmGBPJCW2FCXFck1EoDBF6hks6auNJVQ26M4h73W9W",
				},
				{
					wif2,
					"PtWZ6y56SeRZiuMHBrUkFAbhrURogF7xzWL6PQQJ86XvZfeE3jf1a",
				},
				{
					wif3,
					"PsUVgxwa9RM9m5jBoywyzbP3SjPQ7QC4uNEjUVDsTfBeahKkmETvQ",
				},
				{
					wif4,
					"Pr9H1pVL3qxuXK54u1GRxRpC4VUbEtRdMuG8JT8kcEssBALyTRM7v",
				},
			}
		}

		for _, test := range tests {
			// Test that encoding the WIF structure matches the expected string.
			s := test.wif.String()
			if s != test.encoded {
				t.Errorf("TestEncodeDecodePrivateKey failed: want '%s', got '%s'",
					test.encoded, s)
				continue
			}

			// Test that decoding the expected string results in the original WIF
			// structure.
			w, err := DecodeWIF(test.encoded)
			if err != nil {
				t.Error(err)
				continue
			}
			if got := w.String(); got != test.encoded {
				t.Errorf("NewWIF failed: want '%v', got '%v'", test.wif, got)
			}

			w.SerializePubKey()
		}
	}
}
