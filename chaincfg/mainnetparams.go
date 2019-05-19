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
		{"dnsseed.fonero.org", true},
	},

	// Chain parameters
	GenesisBlock:             &genesisBlock,
	GenesisHash:              &genesisHash,
	PowLimit:                 mainPowLimit,
	PowLimitBits:             0x1d00ffff,
	ReduceMinDifficulty:      true, // false
	MinDiffReductionTime:     time.Minute * 10, // 0 ~99.3% chance to be mined before reduction
	GenerateSupported:        true,
	MaximumBlockSizes:        []int{393216},
	MaxTxSize:                393216,
	TargetTimePerBlock:       time.Minute * 1, // time.Minute * 5
	WorkDiffAlpha:            1,
	WorkDiffWindowSize:       14400, // 144
	WorkDiffWindows:          2000, // 20
	TargetTimespan:           time.Minute * 1 * 14400, // time.Minute * 5 * 144 TimePerBlock * WindowSize
	RetargetAdjustmentFactor: 4,

	// Subsidy parameters.
	BaseSubsidy:              6239165328,
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
	Deployments: map[uint32][]ConsensusDeployment{
		4: {{
			Vote: Vote{
				Id:          VoteIDSDiffAlgorithm,
				Description: "Change stake difficulty algorithm as defined in DCP0001",
				Mask:        0x0006, // Bits 1 and 2
				Choices: []Choice{{
					Id:          "abstain",
					Description: "abstain voting for change",
					Bits:        0x0000,
					IsAbstain:   true,
					IsNo:        false,
				}, {
					Id:          "no",
					Description: "keep the existing algorithm",
					Bits:        0x0002, // Bit 1
					IsAbstain:   false,
					IsNo:        true,
				}, {
					Id:          "yes",
					Description: "change to the new algorithm",
					Bits:        0x0004, // Bit 2
					IsAbstain:   false,
					IsNo:        false,
				}},
			},
			StartTime:  1493164800, // Apr 26th, 2017
			ExpireTime: 1524700800, // Apr 26th, 2018
		}, {
			Vote: Vote{
				Id:          VoteIDLNSupport,
				Description: "Request developers begin work on Lightning Network (LN) integration",
				Mask:        0x0018, // Bits 3 and 4
				Choices: []Choice{{
					Id:          "abstain",
					Description: "abstain from voting",
					Bits:        0x0000,
					IsAbstain:   true,
					IsNo:        false,
				}, {
					Id:          "no",
					Description: "no, do not work on integrating LN support",
					Bits:        0x0008, // Bit 3
					IsAbstain:   false,
					IsNo:        true,
				}, {
					Id:          "yes",
					Description: "yes, begin work on integrating LN support",
					Bits:        0x0010, // Bit 4
					IsAbstain:   false,
					IsNo:        false,
				}},
			},
			StartTime:  1493164800, // Apr 26th, 2017
			ExpireTime: 1508976000, // Oct 26th, 2017
		}},
		5: {{
			Vote: Vote{
				Id:          VoteIDLNFeatures,
				Description: "Enable features defined in DCP0002 and DCP0003 necessary to support Lightning Network (LN)",
				Mask:        0x0006, // Bits 1 and 2
				Choices: []Choice{{
					Id:          "abstain",
					Description: "abstain voting for change",
					Bits:        0x0000,
					IsAbstain:   true,
					IsNo:        false,
				}, {
					Id:          "no",
					Description: "keep the existing consensus rules",
					Bits:        0x0002, // Bit 1
					IsAbstain:   false,
					IsNo:        true,
				}, {
					Id:          "yes",
					Description: "change to the new consensus rules",
					Bits:        0x0004, // Bit 2
					IsAbstain:   false,
					IsNo:        false,
				}},
			},
			StartTime:  1505260800, // Sep 13th, 2017
			ExpireTime: 1536796800, // Sep 13th, 2018
		}},
		6: {{
			Vote: Vote{
				Id:          VoteIDFixLNSeqLocks,
				Description: "Modify sequence lock handling as defined in DCP0004",
				Mask:        0x0006, // Bits 1 and 2
				Choices: []Choice{{
					Id:          "abstain",
					Description: "abstain voting for change",
					Bits:        0x0000,
					IsAbstain:   true,
					IsNo:        false,
				}, {
					Id:          "no",
					Description: "keep the existing consensus rules",
					Bits:        0x0002, // Bit 1
					IsAbstain:   false,
					IsNo:        true,
				}, {
					Id:          "yes",
					Description: "change to the new consensus rules",
					Bits:        0x0004, // Bit 2
					IsAbstain:   false,
					IsNo:        false,
				}},
			},
			StartTime:  1548633600, // Jan 28th, 2019
			ExpireTime: 1580169600, // Jan 28th, 2020
		}},
	},
	
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
	// Organization address is FeUB7x34L2RqHYfFWDgN7unQxo2U5GC6z45
	OrganizationPkScript:        hexDecode("76a9141d7733819d8db24218bf7662193dd4951538108988ac"),
	OrganizationPkScriptVersion: 0,
	BlockOneLedger:              BlockOneLedgerMainNet,
}
