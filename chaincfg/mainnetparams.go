// Copyright (c) 2014-2016 The btcsuite developers
// Copyright (c) 2015-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"time"

	"github.com/fonero-project/fnod/wire"
)

// MainNetParams defines the network parameters for the main Fonero network.
var MainNetParams = Params{
	Name:        "mainnet",
	Net:         wire.MainNet,
	DefaultPort: "9208",
	DNSSeeds: []DNSSeed{
//		{"seed.fonero.org", true},
	},

	// Chain parameters
	GenesisBlock:             &genesisBlock,
	GenesisHash:              &genesisHash,
	PowLimit:                 mainPowLimit,
	PowLimitBits:             0x1d00ffff,
	ReduceMinDifficulty:      false,
	MinDiffReductionTime:     0, // Does not apply since ReduceMinDifficulty false
	GenerateSupported:        true,
	MaximumBlockSizes:        []int{393216},
	MaxTxSize:                393216,
	TargetTimePerBlock:       time.Minute * 5,
	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       144,
	WorkDiffWindows:          20,
	TargetTimespan:           time.Minute * 5 * 144, // TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 4,

	// Subsidy parameters.
	BaseSubsidy:              3119582664, // 21m
	MulSubsidy:               100,
	DivSubsidy:               101,
	SubsidyReductionInterval: 6144,
	WorkRewardProportion:     6,
	StakeRewardProportion:    3,
	BlockTaxProportion:       1,

	// Checkpoints ordered from oldest to newest.
	Checkpoints: []Checkpoint{},

	// The miner confirmation window is defined as:
	//   target proof of work timespan / target proof of work spacing
	RuleChangeActivationQuorum:     4032, // 10 % of RuleChangeActivationInterval * TicketsPerBlock
	RuleChangeActivationMultiplier: 3,    // 75%
	RuleChangeActivationDivisor:    4,
	RuleChangeActivationInterval:   2016 * 4, // 4 weeks

	// Enforce current block version once majority of the network has
	// upgraded.
	// 75% (750 / 1000)
	// Reject previous block versions once a majority of the network has
	// upgraded.
	// 95% (950 / 1000)
	BlockEnforceNumRequired: 750,
	BlockRejectNumRequired:  950,
	BlockUpgradeNumToCheck:  1000,

	// AcceptNonStdTxs is a mempool param to either accept and relay
	// non standard txs to the network or reject them
	AcceptNonStdTxs: false,

	// Address encoding magics
	NetworkAddressPrefix: "F",
	PubKeyAddrID:         [2]byte{0x15, 0x86},
	PubKeyHashAddrID:     [2]byte{0x08, 0x3f},
	PKHEdwardsAddrID:     [2]byte{0x08, 0x1f},
	PKHSchnorrAddrID:     [2]byte{0x08, 0x01},
	ScriptHashAddrID:     [2]byte{0x08, 0x1a},
	PrivateKeyID:         [2]byte{0x23, 0xde},

	// BIP32 hierarchical deterministic extended key magics
	HDPrivateKeyID: [4]byte{0x03, 0xf3, 0xa5, 0xe4},
	HDPublicKeyID:  [4]byte{0x03, 0xf1, 0xa7, 0x16},

	// BIP44 coin type used in the hierarchical deterministic path for
	// address generation.
	SLIP0044CoinType: 42, // SLIP0044, Fonero
	LegacyCoinType:   20, // for backwards compatibility

	// Fonero PoS parameters
	MinimumStakeDiff:        2 * 1e8, // 2 Coin
	TicketPoolSize:          8192,
	TicketsPerBlock:         5,
	TicketMaturity:          256,
	TicketExpiry:            40960, // 5*TicketPoolSize
	CoinbaseMaturity:        256,
	SStxChangeMaturity:      1,
	TicketPoolSizeWeight:    4,
	StakeDiffAlpha:          1, // Minimal
	StakeDiffWindowSize:     144,
	StakeDiffWindows:        20,
	StakeVersionInterval:    144 * 2 * 7, // ~1 week
	MaxFreshStakePerBlock:   20,          // 4*TicketsPerBlock
	StakeEnabledHeight:      256 + 256,   // CoinbaseMaturity + TicketMaturity
	StakeValidationHeight:   4096,        // ~14 days
	StakeBaseSigScript:      []byte{0x00, 0x00},
	StakeMajorityMultiplier: 3,
	StakeMajorityDivisor:    4,

	// Fonero organization related parameters
	// Organization address is Dcur2mcGjmENx4DhNqDctW5wJCVyT3Qeqkx
	OrganizationPkScript:        hexDecode("76a9141c20f907c788443af697e3404805c873afd73dec88ac"),
	OrganizationPkScriptVersion: 0,
	BlockOneLedger:              BlockOneLedgerMainNet,
}
