// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package fnoutil

import (
	"bytes"
	"errors"

	"github.com/decred/base58"
	"github.com/fonero-project/fnod/chaincfg"
	"github.com/fonero-project/fnod/chaincfg/chainec"
	"github.com/fonero-project/fnod/chaincfg/chainhash"
	"github.com/fonero-project/fnod/fnoec"
	"github.com/fonero-project/fnod/fnoec/edwards"
	"github.com/fonero-project/fnod/fnoec/secp256k1"
)

// ErrMalformedPrivateKey describes an error where a WIF-encoded private
// key cannot be decoded due to being improperly formatted.  This may occur
// if the byte length is incorrect or an unexpected magic number was
// encountered.
var ErrMalformedPrivateKey = errors.New("malformed private key")

// WIF contains the individual components described by the Wallet Import Format
// (WIF).  A WIF string is typically used to represent a private key and its
// associated address in a way that  may be easily copied and imported into or
// exported from wallet software.  WIF strings may be decoded into this
// structure by calling DecodeWIF or created with a user-provided private key
// by calling NewWIF.
type WIF struct {
	// ecType is the type of ECDSA used.
	ecType fnoec.SignatureType

	// PrivKey is the private key being imported or exported.
	PrivKey chainec.PrivateKey

	// netID is the network identifier byte used when
	// WIF encoding the private key.
	netID [2]byte
}

// NewWIF creates a new WIF structure to export an address and its private key
// as a string encoded in the Wallet Import Format.  The compress argument
// specifies whether the address intended to be imported or exported was created
// by serializing the public key compressed rather than uncompressed.
func NewWIF(privKey chainec.PrivateKey, net *chaincfg.Params, ecType fnoec.SignatureType) (*WIF,
	error) {
	if net == nil {
		return nil, errors.New("no network")
	}
	return &WIF{ecType, privKey, net.PrivateKeyID}, nil
}

// IsForNet returns whether or not the decoded WIF structure is associated
// with the passed network.
func (w *WIF) IsForNet(net *chaincfg.Params) bool {
	return w.netID == net.PrivateKeyID
}

// DecodeWIF creates a new WIF structure by decoding the string encoding of
// the import format.
//
// The WIF string must be a base58-encoded string of the following byte
// sequence:
//
//  * 2 bytes to identify the network, must be 0x80 for mainnet or 0xef for testnet
//  * 1 byte for ECDSA type
//  * 32 bytes of a binary-encoded, big-endian, zero-padded private key
//  * 4 bytes of checksum, must equal the first four bytes of the double SHA256
//    of every byte before the checksum in this sequence
//
// If the base58-decoded byte sequence does not match this, DecodeWIF will
// return a non-nil error.  ErrMalformedPrivateKey is returned when the WIF
// is of an impossible length.  ErrChecksumMismatch is returned if the
// expected WIF checksum does not match the calculated checksum.
func DecodeWIF(wif string) (*WIF, error) {
	decoded := base58.Decode(wif)
	decodedLen := len(decoded)

	if decodedLen != 39 {
		return nil, ErrMalformedPrivateKey
	}

	// Checksum is first four bytes of hash of the identifier byte
	// and privKey.  Verify this matches the final 4 bytes of the decoded
	// private key.
	cksum := chainhash.HashB(decoded[:decodedLen-4])
	if !bytes.Equal(cksum[:4], decoded[decodedLen-4:]) {
		return nil, ErrChecksumMismatch
	}

	netID := [2]byte{decoded[0], decoded[1]}
	var privKey chainec.PrivateKey

	var ecType fnoec.SignatureType
	switch fnoec.SignatureType(decoded[2]) {
	case fnoec.STEcdsaSecp256k1:
		privKeyBytes := decoded[3 : 3+secp256k1.PrivKeyBytesLen]
		privKey, _ = secp256k1.PrivKeyFromScalar(privKeyBytes)
		ecType = fnoec.STEcdsaSecp256k1
	case fnoec.STEd25519:
		privKeyBytes := decoded[3 : 3+edwards.PrivScalarSize]
		privKey, _, _ = edwards.PrivKeyFromScalar(edwards.Edwards(), privKeyBytes)
		ecType = fnoec.STEd25519
	case fnoec.STSchnorrSecp256k1:
		privKeyBytes := decoded[3 : 3+secp256k1.PrivKeyBytesLen]
		privKey, _ = secp256k1.PrivKeyFromScalar(privKeyBytes)
		ecType = fnoec.STSchnorrSecp256k1
	}

	return &WIF{ecType, privKey, netID}, nil
}

// String creates the Wallet Import Format string encoding of a WIF structure.
// See DecodeWIF for a detailed breakdown of the format and requirements of
// a valid WIF string.
func (w *WIF) String() string {
	// Precalculate size.  Maximum number of bytes before base58 encoding
	// is two bytes for the network, one byte for the ECDSA type, 32 bytes
	// of private key and finally four bytes of checksum.
	encodeLen := 2 + 1 + 32 + 4

	a := make([]byte, 0, encodeLen)
	a = append(a, w.netID[:]...)
	a = append(a, byte(w.ecType))
	a = append(a, w.PrivKey.Serialize()...)

	cksum := chainhash.HashB(a)
	a = append(a, cksum[:4]...)
	return base58.Encode(a)
}

// SerializePubKey serializes the associated public key of the imported or
// exported private key in compressed format.  The serialization format
// chosen depends on the value of w.ecType.
func (w *WIF) SerializePubKey() []byte {
	pkx, pky := w.PrivKey.Public()
	var pk chainec.PublicKey

	switch w.ecType {
	case fnoec.STEcdsaSecp256k1:
		pk = secp256k1.NewPublicKey(pkx, pky)
	case fnoec.STEd25519:
		pk = edwards.NewPublicKey(edwards.Edwards(), pkx, pky)
	case fnoec.STSchnorrSecp256k1:
		pk = secp256k1.NewPublicKey(pkx, pky)
	}

	return pk.SerializeCompressed()
}

// DSA returns the ECDSA type for the private key.
func (w *WIF) DSA() fnoec.SignatureType {
	return w.ecType
}
