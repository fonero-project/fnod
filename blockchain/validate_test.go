// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2018 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package blockchain

import (
	"bytes"
	"compress/bzip2"
	"encoding/gob"
	"fmt"
	mrand "math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fonero-project/fnod/blockchain/chaingen"
	"github.com/fonero-project/fnod/chaincfg"
	"github.com/fonero-project/fnod/chaincfg/chainhash"
	"github.com/fonero-project/fnod/database"
	"github.com/fonero-project/fnod/fnoutil"
	"github.com/fonero-project/fnod/wire"
)

// TestBlockchainSpendJournal tests for whether or not the spend journal is being
// written to disk correctly on a live blockchain.
func TestBlockchainSpendJournal(t *testing.T) {
	// Update parameters to reflect what is expected by the legacy data.
	params := cloneParams(&chaincfg.RegNetParams)
	params.GenesisBlock.Header.MerkleRoot = *mustParseHash("a216ea043f0d481a072424af646787794c32bcefd3ed181a090319bbf8a37105")
	params.GenesisBlock.Header.Timestamp = time.Unix(1401292357, 0)
	params.GenesisBlock.Transactions[0].TxIn[0].ValueIn = 0
	params.PubKeyHashAddrID = [2]byte{0x0e, 0x91}
	params.StakeBaseSigScript = []byte{0xde, 0xad, 0xbe, 0xef}
	params.OrganizationPkScript = hexToBytes("a914cbb08d6ca783b533b2c7d24a51fbca92d937bf9987")
	params.BlockOneLedger = []*chaincfg.TokenPayout{
		{Address: "Sshw6S86G2bV6W32cbc7EhtFy8f93rU6pae", Amount: 100000 * 1e8},
		{Address: "SsjXRK6Xz6CFuBt6PugBvrkdAa4xGbcZ18w", Amount: 100000 * 1e8},
		{Address: "SsfXiYkYkCoo31CuVQw428N6wWKus2ZEw5X", Amount: 100000 * 1e8},
	}
	genesisHash := params.GenesisBlock.BlockHash()
	params.GenesisHash = &genesisHash

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("spendjournalunittest", params)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// Load up the rest of the blocks up to HEAD.
	filename := filepath.Join("testdata", "reorgto179.bz2")
	fi, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open %s: %v", filename, err)
	}
	bcStream := bzip2.NewReader(fi)
	defer fi.Close()

	// Create a buffer of the read file
	bcBuf := new(bytes.Buffer)
	bcBuf.ReadFrom(bcStream)

	// Create decoder from the buffer and a map to store the data
	bcDecoder := gob.NewDecoder(bcBuf)
	blockChain := make(map[int64][]byte)

	// Decode the blockchain into the map
	if err := bcDecoder.Decode(&blockChain); err != nil {
		t.Errorf("error decoding test blockchain: %v", err.Error())
	}

	// Load up the short chain
	finalIdx1 := 179
	for i := 1; i < finalIdx1+1; i++ {
		bl, err := fnoutil.NewBlockFromBytes(blockChain[int64(i)])
		if err != nil {
			t.Fatalf("NewBlockFromBytes error: %v", err.Error())
		}

		forkLen, isOrphan, err := chain.ProcessBlock(bl, BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock error at height %v: %v", i, err.Error())
		}
		isMainChain := !isOrphan && forkLen == 0
		if !isMainChain {
			t.Fatalf("block %s (height %d) should have been "+
				"accepted to the main chain", bl.Hash(),
				bl.MsgBlock().Header.Height)
		}
	}

	// Loop through all of the blocks and ensure the number of spent outputs
	// matches up with the information loaded from the spend journal.
	err = chain.db.View(func(dbTx database.Tx) error {
		for i := int64(2); i <= chain.bestChain.Tip().height; i++ {
			node := chain.bestChain.NodeByHeight(i)
			if node == nil {
				str := fmt.Sprintf("no block at height %d exists", i)
				return errNotInMainChain(str)
			}
			block, err := dbFetchBlockByNode(dbTx, node)
			if err != nil {
				return err
			}

			ntx := countSpentOutputs(block)
			stxos, err := dbFetchSpendJournalEntry(dbTx, block)
			if err != nil {
				return err
			}

			if ntx != len(stxos) {
				return fmt.Errorf("bad number of stxos "+
					"calculated at "+"height %v, got %v "+
					"expected %v", i, len(stxos), ntx)
			}
		}

		return nil
	})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestSequenceLocksActive ensure the sequence locks are detected as active or
// not as expected in all possible scenarios.
func TestSequenceLocksActive(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name          string
		seqLockHeight int64
		seqLockTime   int64
		blockHeight   int64
		medianTime    int64
		want          bool
	}{
		{
			// Block based sequence lock with height at min
			// required.
			name:          "min active block height",
			seqLockHeight: 1000,
			seqLockTime:   -1,
			blockHeight:   1001,
			medianTime:    now + 31,
			want:          true,
		},
		{
			// Time based sequence lock with relative time at min
			// required.
			name:          "min active median time",
			seqLockHeight: -1,
			seqLockTime:   now + 30,
			blockHeight:   1001,
			medianTime:    now + 31,
			want:          true,
		},
		{
			// Block based sequence lock at same height.
			name:          "same height",
			seqLockHeight: 1000,
			seqLockTime:   -1,
			blockHeight:   1000,
			medianTime:    now + 31,
			want:          false,
		},
		{
			// Time based sequence lock with relative time equal to
			// lock time.
			name:          "same median time",
			seqLockHeight: -1,
			seqLockTime:   now + 30,
			blockHeight:   1001,
			medianTime:    now + 30,
			want:          false,
		},
		{
			// Block based sequence lock with relative height below
			// required.
			name:          "height below required",
			seqLockHeight: 1000,
			seqLockTime:   -1,
			blockHeight:   999,
			medianTime:    now + 31,
			want:          false,
		},
		{
			// Time based sequence lock with relative time before
			// required.
			name:          "median time before required",
			seqLockHeight: -1,
			seqLockTime:   now + 30,
			blockHeight:   1001,
			medianTime:    now + 29,
			want:          false,
		},
	}

	for _, test := range tests {
		seqLock := SequenceLock{
			MinHeight: test.seqLockHeight,
			MinTime:   test.seqLockTime,
		}
		got := SequenceLockActive(&seqLock, test.blockHeight,
			time.Unix(test.medianTime, 0))
		if got != test.want {
			t.Errorf("%s: mismatched seqence lock status - got %v, "+
				"want %v", test.name, got, test.want)
			continue
		}
	}
}

// quickVoteActivationParams returns a set of test chain parameters which allow
// for quicker vote activation as compared to various existing network params by
// reducing the required maturities, the ticket pool size, the stake enabled and
// validation heights, the proof-of-work block version upgrade window, the stake
// version interval, and the rule change activation interval.
func quickVoteActivationParams() *chaincfg.Params {
	params := cloneParams(&chaincfg.RegNetParams)
	params.WorkDiffWindowSize = 200000
	params.WorkDiffWindows = 1
	params.TargetTimespan = params.TargetTimePerBlock *
		time.Duration(params.WorkDiffWindowSize)
	params.CoinbaseMaturity = 2
	params.BlockEnforceNumRequired = 5
	params.BlockRejectNumRequired = 7
	params.BlockUpgradeNumToCheck = 10
	params.TicketMaturity = 2
	params.TicketPoolSize = 4
	params.TicketExpiry = 6 * uint32(params.TicketPoolSize)
	params.StakeEnabledHeight = int64(params.CoinbaseMaturity) +
		int64(params.TicketMaturity)
	params.StakeValidationHeight = int64(params.CoinbaseMaturity) +
		int64(params.TicketPoolSize)*2
	params.StakeVersionInterval = 10
	params.RuleChangeActivationInterval = uint32(params.TicketPoolSize) *
		uint32(params.TicketsPerBlock)
	params.RuleChangeActivationQuorum = params.RuleChangeActivationInterval *
		uint32(params.TicketsPerBlock*100) / 1000

	return params
}

// TestLegacySequenceLocks ensure that sequence locks within blocks behave as
// expected according to the legacy semantics in previous version of the
// software.
func TestLegacySequenceLocks(t *testing.T) {
	// Use a set of test chain parameters which allow for quicker vote
	// activation as compared to various existing network params.
	params := quickVoteActivationParams()

	// deploymentVer is the deployment version of the ln features vote for the
	// chain params.
	const deploymentVer = 6

	// Find the correct deployment for the LN features agenda.
	voteID := chaincfg.VoteIDLNFeatures
	var deployment *chaincfg.ConsensusDeployment
	deployments := params.Deployments[deploymentVer]
	for deploymentID, depl := range deployments {
		if depl.Vote.Id == voteID {
			deployment = &deployments[deploymentID]
			break
		}
	}
	if deployment == nil {
		t.Fatalf("Unable to find consensus deployement for %s", voteID)
	}

	// Find the correct choice for the yes vote.
	const yesVoteID = "yes"
	var yesChoice *chaincfg.Choice
	for _, choice := range deployment.Vote.Choices {
		if choice.Id == yesVoteID {
			yesChoice = &choice
		}
	}
	if yesChoice == nil {
		t.Fatalf("Unable to find vote choice for id %q", yesVoteID)
	}

	// Create a test generator instance initialized with the genesis block as
	// the tip.
	g, err := chaingen.MakeGenerator(params)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("seqlocksoldsemanticstest", params)
	if err != nil {
		t.Fatalf("Failed to setup chain instance: %v", err)
	}
	defer teardownFunc()

	// accepted processes the current tip block associated with the generator
	// and expects it to be accepted to the main chain.
	//
	// rejected expects the block to be rejected with the provided error code.
	//
	// expectTip expects the provided block to be the current tip of the
	// main chain.
	//
	// acceptedToSideChainWithExpectedTip expects the block to be accepted to a
	// side chain, but the current best chain tip to be the provided value.
	//
	// testThresholdState queries the threshold state from the current tip block
	// associated with the generator and expects the returned state to match the
	// provided value.
	accepted := func() {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)", g.TipName(),
			block.Hash(), blockHeight)

		forkLen, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should have been "+
				"accepted: %v", g.TipName(), block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values specified in
		// the test.
		isMainChain := !isOrphan && forkLen == 0
		if !isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main chain "+
				"flag -- got %v, want true", g.TipName(), block.Hash(),
				blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected orphan flag -- "+
				"got %v, want false", g.TipName(), block.Hash(), blockHeight,
				isOrphan)
		}
	}
	rejected := func(code ErrorCode) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)", g.TipName(),
			block.Hash(), blockHeight)

		_, _, err := chain.ProcessBlock(block, BFNone)
		if err == nil {
			t.Fatalf("block %q (hash %s, height %d) should not have been "+
				"accepted", g.TipName(), block.Hash(), blockHeight)
		}

		// Ensure the error code is of the expected type and the reject code
		// matches the value specified in the test instance.
		rerr, ok := err.(RuleError)
		if !ok {
			t.Fatalf("block %q (hash %s, height %d) returned unexpected error "+
				"type -- got %T, want blockchain.RuleError", g.TipName(),
				block.Hash(), blockHeight, err)
		}
		if rerr.ErrorCode != code {
			t.Fatalf("block %q (hash %s, height %d) does not have expected "+
				"reject code -- got %v, want %v", g.TipName(), block.Hash(),
				blockHeight, rerr.ErrorCode, code)
		}
	}
	expectTip := func(tipName string) {
		// Ensure hash and height match.
		wantTip := g.BlockByName(tipName)
		best := chain.BestSnapshot()
		if best.Hash != wantTip.BlockHash() ||
			best.Height != int64(wantTip.Header.Height) {
			t.Fatalf("block %q (hash %s, height %d) should be the current tip "+
				"-- got (hash %s, height %d)", tipName, wantTip.BlockHash(),
				wantTip.Header.Height, best.Hash, best.Height)
		}
	}
	acceptedToSideChainWithExpectedTip := func(tipName string) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)", g.TipName(),
			block.Hash(), blockHeight)

		forkLen, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should have been "+
				"accepted: %v", g.TipName(), block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values specified in
		// the test.
		isMainChain := !isOrphan && forkLen == 0
		if isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main chain "+
				"flag -- got %v, want false", g.TipName(), block.Hash(),
				blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected orphan flag -- "+
				"got %v, want false", g.TipName(), block.Hash(), blockHeight,
				isOrphan)
		}

		expectTip(tipName)
	}
	testThresholdState := func(id string, state ThresholdState) {
		tipHash := g.Tip().BlockHash()
		s, err := chain.NextThresholdState(&tipHash, deploymentVer, id)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) unexpected error when "+
				"retrieving threshold state: %v", g.TipName(), tipHash,
				g.Tip().Header.Height, err)
		}

		if s.State != state {
			t.Fatalf("block %q (hash %s, height %d) unexpected threshold "+
				"state for %s -- got %v, want %v", g.TipName(), tipHash,
				g.Tip().Header.Height, id, s.State, state)
		}
	}

	// Shorter versions of useful params for convenience.
	ticketsPerBlock := int64(params.TicketsPerBlock)
	coinbaseMaturity := params.CoinbaseMaturity
	stakeEnabledHeight := params.StakeEnabledHeight
	stakeValidationHeight := params.StakeValidationHeight
	stakeVerInterval := params.StakeVersionInterval
	ruleChangeInterval := int64(params.RuleChangeActivationInterval)

	// ---------------------------------------------------------------------
	// First block.
	// ---------------------------------------------------------------------

	// Add the required first block.
	//
	//   genesis -> bp
	g.CreatePremineBlock("bp", 0)
	g.AssertTipHeight(1)
	accepted()

	// ---------------------------------------------------------------------
	// Generate enough blocks to have mature coinbase outputs to work with.
	//
	//   genesis -> bp -> bm0 -> bm1 -> ... -> bm#
	// ---------------------------------------------------------------------

	for i := uint16(0); i < coinbaseMaturity; i++ {
		blockName := fmt.Sprintf("bm%d", i)
		g.NextBlock(blockName, nil, nil)
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(coinbaseMaturity) + 1)

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake enabled height while
	// creating ticket purchases that spend from the coinbases matured
	// above.  This will also populate the pool of immature tickets.
	//
	//   ... -> bm# ... -> bse0 -> bse1 -> ... -> bse#
	// ---------------------------------------------------------------------

	var ticketsPurchased int
	for i := int64(0); int64(g.Tip().Header.Height) < stakeEnabledHeight; i++ {
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		ticketsPurchased += len(ticketOuts)
		blockName := fmt.Sprintf("bse%d", i)
		g.NextBlock(blockName, nil, ticketOuts)
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeEnabledHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake validation height while
	// continuing to purchase tickets using the coinbases matured above and
	// allowing the immature tickets to mature and thus become live.
	//
	// The blocks are also generated with version 6 to ensure stake version
	// and ln feature enforcement is reached.
	//
	//   ... -> bse# -> bsv0 -> bsv1 -> ... -> bsv#
	// ---------------------------------------------------------------------

	targetPoolSize := int64(g.Params().TicketPoolSize) * ticketsPerBlock
	for i := int64(0); int64(g.Tip().Header.Height) < stakeValidationHeight; i++ {
		// Only purchase tickets until the target ticket pool size is reached.
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		if ticketsPurchased+len(ticketOuts) > int(targetPoolSize) {
			ticketsNeeded := int(targetPoolSize) - ticketsPurchased
			if ticketsNeeded > 0 {
				ticketOuts = ticketOuts[1 : ticketsNeeded+1]
			} else {
				ticketOuts = nil
			}
		}
		ticketsPurchased += len(ticketOuts)

		blockName := fmt.Sprintf("bsv%d", i)
		g.NextBlock(blockName, nil, ticketOuts,
			chaingen.ReplaceBlockVersion(6))
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach one block before the next two stake
	// version intervals with block version 6, stake version 0, and vote
	// version 6.
	//
	// This will result in triggering enforcement of the stake version and
	// that the stake version is 6.  The treshold state for deployment will
	// move to started since the next block also coincides with the start of
	// a new rule change activation interval for the chosen parameters.
	//
	//   ... -> bsv# -> bvu0 -> bvu1 -> ... -> bvu#
	// ---------------------------------------------------------------------

	blocksNeeded := stakeValidationHeight + stakeVerInterval*2 - 1 -
		int64(g.Tip().Header.Height)
	for i := int64(0); i < blocksNeeded; i++ {
		outs := g.OldestCoinbaseOuts()
		blockName := fmt.Sprintf("bvu%d", i)
		g.NextBlock(blockName, nil, outs[1:], chaingen.ReplaceBlockVersion(6),
			chaingen.ReplaceVoteVersions(6))
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	testThresholdState(voteID, ThresholdStarted)

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the next rule change interval with
	// block version 6, stake version 6, and vote version 6.  Also, set the
	// vote bits to include yes votes for the ln feature agenda.
	//
	// This will result in moving the threshold state for the ln feature
	// enforcement to locked in.
	//
	//   ... -> bvu# -> bvli0 -> bvli1 -> ... -> bvli#
	// ---------------------------------------------------------------------

	blocksNeeded = stakeValidationHeight + ruleChangeInterval*2 - 1 -
		int64(g.Tip().Header.Height)
	for i := int64(0); i < blocksNeeded; i++ {
		outs := g.OldestCoinbaseOuts()
		blockName := fmt.Sprintf("bvli%d", i)
		g.NextBlock(blockName, nil, outs[1:], chaingen.ReplaceBlockVersion(6),
			chaingen.ReplaceStakeVersion(6),
			chaingen.ReplaceVotes(vbPrevBlockValid|yesChoice.Bits, 6))
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight + ruleChangeInterval*2 - 1))
	g.AssertBlockVersion(6)
	g.AssertStakeVersion(6)
	testThresholdState(voteID, ThresholdLockedIn)

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the next rule change interval with
	// block version 6, stake version 6, and vote version 6.
	//
	// This will result in moving the threshold state for the ln feature
	// enforcement to active thereby activating it.
	//
	//   ... -> bvli# -> bva0 -> bva1 -> ... -> bva#
	// ---------------------------------------------------------------------

	blocksNeeded = stakeValidationHeight + ruleChangeInterval*3 - 1 -
		int64(g.Tip().Header.Height)
	for i := int64(0); i < blocksNeeded; i++ {
		outs := g.OldestCoinbaseOuts()
		blockName := fmt.Sprintf("bva%d", i)
		g.NextBlock(blockName, nil, outs[1:], chaingen.ReplaceBlockVersion(6),
			chaingen.ReplaceStakeVersion(6), chaingen.ReplaceVoteVersions(6))
		g.SaveTipCoinbaseOuts()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight + ruleChangeInterval*3 - 1))
	g.AssertBlockVersion(6)
	g.AssertStakeVersion(6)
	testThresholdState(voteID, ThresholdActive)

	// ---------------------------------------------------------------------
	// Perform a series of sequence lock tests now that ln feature
	// enforcement is active.
	// ---------------------------------------------------------------------

	// enableSeqLocks modifies the passed transaction to enable sequence locks
	// for the provided input.
	enableSeqLocks := func(tx *wire.MsgTx, txInIdx int) {
		tx.Version = 2
		tx.TxIn[txInIdx].Sequence = 0
	}

	// ---------------------------------------------------------------------
	// Create block that has a transaction with an input shared with a
	// transaction in the stake tree and has several outputs used in
	// subsequent blocks.  Also, enable sequence locks for the first of
	// those outputs.
	//
	//   ... -> b0
	// ---------------------------------------------------------------------

	outs := g.OldestCoinbaseOuts()
	b0 := g.NextBlock("b0", &outs[0], outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			// Save the current outputs of the spend tx and clear them.
			tx := b.Transactions[1]
			origOut := tx.TxOut[0]
			origOpReturnOut := tx.TxOut[1]
			tx.TxOut = tx.TxOut[:0]

			// Evenly split the original output amount over multiple outputs.
			const numOutputs = 6
			amount := origOut.Value / numOutputs
			for i := 0; i < numOutputs; i++ {
				if i == numOutputs-1 {
					amount = origOut.Value - amount*(numOutputs-1)
				}
				tx.AddTxOut(wire.NewTxOut(int64(amount), origOut.PkScript))
			}

			// Add the original op return back to the outputs and enable
			// sequence locks for the first output.
			tx.AddTxOut(origOpReturnOut)
			enableSeqLocks(tx, 0)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	// ---------------------------------------------------------------------
	// Create block that spends from an output created in the previous
	// block.
	//
	//   ... -> b0 -> b1a
	// ---------------------------------------------------------------------

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b1a", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 0)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	accepted()

	// ---------------------------------------------------------------------
	// Create block that involves reorganize to a sequence lock spending
	// from an output created in a block prior to the parent also spent on
	// on the side chain.
	//
	//   ... -> b0 -> b1  -> b2
	//            \-> b1a
	// ---------------------------------------------------------------------
	g.SetTip("b0")
	g.NextBlock("b1", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6))
	g.SaveTipCoinbaseOuts()
	acceptedToSideChainWithExpectedTip("b1a")

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b2", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 0)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	g.SaveTipCoinbaseOuts()
	accepted()
	expectTip("b2")

	// ---------------------------------------------------------------------
	// Create block that involves a sequence lock on a vote.
	//
	//   ... -> b2 -> b3
	// ---------------------------------------------------------------------

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b3", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			enableSeqLocks(b.STransactions[0], 0)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	// ---------------------------------------------------------------------
	// Create block that involves a sequence lock on a ticket.
	//
	//   ... -> b3 -> b4
	// ---------------------------------------------------------------------

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b4", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			enableSeqLocks(b.STransactions[5], 0)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	// ---------------------------------------------------------------------
	// Create two blocks such that the tip block involves a sequence lock
	// spending from a different output of a transaction the parent block
	// also spends from.
	//
	//   ... -> b4 -> b5 -> b6
	// ---------------------------------------------------------------------

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b5", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 1)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			b.AddTransaction(tx)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b6", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 2)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	// ---------------------------------------------------------------------
	// Create block that involves a sequence lock spending from a regular
	// tree transaction earlier in the block.  It should be rejected due
	// to a consensus bug.
	//
	//   ... -> b6
	//            \-> b7
	// ---------------------------------------------------------------------

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b7", &outs[0], outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b, 1, 0)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	rejected(ErrMissingTxOut)

	// ---------------------------------------------------------------------
	// Create block that involves a sequence lock spending from a block
	// prior to the parent.  It should be rejected due to a consensus bug.
	//
	//   ... -> b6 -> b8
	//                  \-> b9
	// ---------------------------------------------------------------------

	g.SetTip("b6")
	g.NextBlock("b8", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6))
	g.SaveTipCoinbaseOuts()
	accepted()

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b9", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 3)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	rejected(ErrMissingTxOut)

	// ---------------------------------------------------------------------
	// Create two blocks such that the tip block involves a sequence lock
	// spending from a different output of a transaction the parent block
	// also spends from when the parent block has been disapproved.    It
	// should be rejected due to a consensus bug.
	//
	//   ... -> b8 -> b10
	//                   \-> b11
	// ---------------------------------------------------------------------

	const (
		// vbDisapprovePrev and vbApprovePrev represent no and yes votes,
		// respectively, on whether or not to approve the previous block.
		vbDisapprovePrev = 0x0000
		vbApprovePrev    = 0x0001
	)

	g.SetTip("b8")
	g.NextBlock("b10", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVoteVersions(6), func(b *wire.MsgBlock) {
			spend := chaingen.MakeSpendableOut(b0, 1, 4)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			b.AddTransaction(tx)
		})
	g.SaveTipCoinbaseOuts()
	accepted()

	outs = g.OldestCoinbaseOuts()
	g.NextBlock("b11", nil, outs[1:],
		chaingen.ReplaceBlockVersion(6), chaingen.ReplaceStakeVersion(6),
		chaingen.ReplaceVotes(vbDisapprovePrev, 6), func(b *wire.MsgBlock) {
			b.Header.VoteBits &^= vbApprovePrev
			spend := chaingen.MakeSpendableOut(b0, 1, 5)
			tx := g.CreateSpendTx(&spend, fnoutil.Amount(1))
			enableSeqLocks(tx, 0)
			b.AddTransaction(tx)
		})
	rejected(ErrMissingTxOut)
}

// TestCheckBlockSanity tests the context free block sanity checks with blocks
// not on a chain.
func TestCheckBlockSanity(t *testing.T) {
	params := &chaincfg.RegNetParams
	timeSource := NewMedianTime()
	block := fnoutil.NewBlock(&badBlock)
	err := CheckBlockSanity(block, timeSource, params)
	if err == nil {
		t.Fatalf("block should fail.\n")
	}
}

// TestCheckBlockHeaderContext tests that genesis block passes context headers
// because its parent is nil.
func TestCheckBlockHeaderContext(t *testing.T) {
	// Create a new database for the blocks.
	params := &chaincfg.RegNetParams
	dbPath := filepath.Join(os.TempDir(), "examplecheckheadercontext")
	_ = os.RemoveAll(dbPath)
	db, err := database.Create("ffldb", dbPath, params.Net)
	if err != nil {
		t.Fatalf("Failed to create database: %v\n", err)
		return
	}
	defer os.RemoveAll(dbPath)
	defer db.Close()

	// Create a new BlockChain instance using the underlying database for
	// the simnet network.
	chain, err := New(&Config{
		DB:          db,
		ChainParams: params,
		TimeSource:  NewMedianTime(),
	})
	if err != nil {
		t.Fatalf("Failed to create chain instance: %v\n", err)
		return
	}

	err = chain.checkBlockHeaderContext(&params.GenesisBlock.Header, nil, BFNone)
	if err != nil {
		t.Fatalf("genesisblock should pass just by definition: %v\n", err)
		return
	}

	// Test failing checkBlockHeaderContext when calcNextRequiredDifficulty
	// fails.
	block := fnoutil.NewBlock(&badBlock)
	newNode := newBlockNode(&block.MsgBlock().Header, nil)
	err = chain.checkBlockHeaderContext(&block.MsgBlock().Header, newNode, BFNone)
	if err == nil {
		t.Fatalf("Should fail due to bad diff in newNode\n")
		return
	}
}

// TestTxValidationErrors ensures certain malformed freestanding transactions
// are rejected as as expected.
func TestTxValidationErrors(t *testing.T) {
	// Create a transaction that is too large
	tx := wire.NewMsgTx()
	prevOut := wire.NewOutPoint(&chainhash.Hash{0x01}, 0, wire.TxTreeRegular)
	tx.AddTxIn(wire.NewTxIn(prevOut, 0, nil))
	pkScript := bytes.Repeat([]byte{0x00}, wire.MaxBlockPayload)
	tx.AddTxOut(wire.NewTxOut(0, pkScript))

	// Assert the transaction is larger than the max allowed size.
	txSize := tx.SerializeSize()
	if txSize <= wire.MaxBlockPayload {
		t.Fatalf("generated transaction is not large enough -- got "+
			"%d, want > %d", txSize, wire.MaxBlockPayload)
	}

	// Ensure transaction is rejected due to being too large.
	err := CheckTransactionSanity(tx, &chaincfg.MainNetParams)
	rerr, ok := err.(RuleError)
	if !ok {
		t.Fatalf("CheckTransactionSanity: unexpected error type for "+
			"transaction that is too large -- got %T", err)
	}
	if rerr.ErrorCode != ErrTxTooBig {
		t.Fatalf("CheckTransactionSanity: unexpected error code for "+
			"transaction that is too large -- got %v, want %v",
			rerr.ErrorCode, ErrTxTooBig)
	}
}

// badBlock is an intentionally bad block that should fail the context-less
// sanity checks.
var badBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:      1,
		MerkleRoot:   chaincfg.RegNetParams.GenesisBlock.Header.MerkleRoot,
		VoteBits:     uint16(0x0000),
		FinalState:   [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Voters:       uint16(0x0000),
		FreshStake:   uint8(0x00),
		Revocations:  uint8(0x00),
		Timestamp:    time.Unix(1401292357, 0), // 2009-01-08 20:54:25 -0600 CST
		PoolSize:     uint32(0),
		Bits:         0x207fffff, // 545259519
		SBits:        int64(0x0000000000000000),
		Nonce:        0x37580963,
		StakeVersion: uint32(0),
		Height:       uint32(0),
	},
	Transactions:  []*wire.MsgTx{},
	STransactions: []*wire.MsgTx{},
}

// TestCheckConnectBlockTemplate ensures that the code which deals with
// checking block templates works as expected.
func TestCheckConnectBlockTemplate(t *testing.T) {
	// Create a test generator instance initialized with the genesis block
	// as the tip as well as some cached payment scripts to be used
	// throughout the tests.
	params := &chaincfg.RegNetParams
	g, err := chaingen.MakeGenerator(params)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	// Create a new database and chain instance to run tests against.
	chain, teardownFunc, err := chainSetup("connectblktemplatetest", params)
	if err != nil {
		t.Fatalf("Failed to setup chain instance: %v", err)
	}
	defer teardownFunc()

	// Define some convenience helper functions to process the current tip
	// block associated with the generator.
	//
	// accepted expects the block to be accepted to the main chain.
	//
	// rejected expects the block to be rejected with the provided error
	// code.
	//
	// expectTip expects the provided block to be the current tip of the
	// main chain.
	//
	// acceptedToSideChainWithExpectedTip expects the block to be accepted
	// to a side chain, but the current best chain tip to be the provided
	// value.
	//
	// forceTipReorg forces the chain instance to reorganize the current tip
	// of the main chain from the given block to the given block.  An error
	// will result if the provided from block is not actually the current
	// tip.
	//
	// acceptedBlockTemplate expected the block to considered a valid block
	// template.
	//
	// rejectedBlockTemplate expects the block to be considered an invalid
	// block template due to the provided error code.
	accepted := func() {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		forkLen, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should "+
				"have been accepted: %v", g.TipName(),
				block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values
		// specified in the test.
		isMainChain := !isOrphan && forkLen == 0
		if !isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main "+
				"chain flag -- got %v, want true", g.TipName(),
				block.Hash(), blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected "+
				"orphan flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isOrphan)
		}
	}
	rejected := func(code ErrorCode) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)", g.TipName(),
			block.Hash(), blockHeight)

		_, _, err := chain.ProcessBlock(block, BFNone)
		if err == nil {
			t.Fatalf("block %q (hash %s, height %d) should not "+
				"have been accepted", g.TipName(), block.Hash(),
				blockHeight)
		}

		// Ensure the error code is of the expected type and the reject
		// code matches the value specified in the test instance.
		rerr, ok := err.(RuleError)
		if !ok {
			t.Fatalf("block %q (hash %s, height %d) returned "+
				"unexpected error type -- got %T, want "+
				"blockchain.RuleError", g.TipName(),
				block.Hash(), blockHeight, err)
		}
		if rerr.ErrorCode != code {
			t.Fatalf("block %q (hash %s, height %d) does not have "+
				"expected reject code -- got %v, want %v",
				g.TipName(), block.Hash(), blockHeight,
				rerr.ErrorCode, code)
		}
	}
	expectTip := func(tipName string) {
		// Ensure hash and height match.
		wantTip := g.BlockByName(tipName)
		best := chain.BestSnapshot()
		if best.Hash != wantTip.BlockHash() ||
			best.Height != int64(wantTip.Header.Height) {
			t.Fatalf("block %q (hash %s, height %d) should be "+
				"the current tip -- got (hash %s, height %d)",
				tipName, wantTip.BlockHash(),
				wantTip.Header.Height, best.Hash, best.Height)
		}
	}
	acceptedToSideChainWithExpectedTip := func(tipName string) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		forkLen, isOrphan, err := chain.ProcessBlock(block, BFNone)
		if err != nil {
			t.Fatalf("block %q (hash %s, height %d) should "+
				"have been accepted: %v", g.TipName(),
				block.Hash(), blockHeight, err)
		}

		// Ensure the main chain and orphan flags match the values
		// specified in the test.
		isMainChain := !isOrphan && forkLen == 0
		if isMainChain {
			t.Fatalf("block %q (hash %s, height %d) unexpected main "+
				"chain flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isMainChain)
		}
		if isOrphan {
			t.Fatalf("block %q (hash %s, height %d) unexpected "+
				"orphan flag -- got %v, want false", g.TipName(),
				block.Hash(), blockHeight, isOrphan)
		}

		expectTip(tipName)
	}
	forceTipReorg := func(fromTipName, toTipName string) {
		from := g.BlockByName(fromTipName)
		to := g.BlockByName(toTipName)
		err = chain.ForceHeadReorganization(from.BlockHash(), to.BlockHash())
		if err != nil {
			t.Fatalf("failed to force header reorg from block %q "+
				"(hash %s, height %d) to block %q (hash %s, "+
				"height %d): %v", fromTipName, from.BlockHash(),
				from.Header.Height, toTipName, to.BlockHash(),
				to.Header.Height, err)
		}
	}
	acceptedBlockTemplate := func() {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block template %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		err := chain.CheckConnectBlockTemplate(block)
		if err != nil {
			t.Fatalf("block template %q (hash %s, height %d) should "+
				"have been accepted: %v", g.TipName(),
				block.Hash(), blockHeight, err)
		}
	}
	rejectedBlockTemplate := func(code ErrorCode) {
		msgBlock := g.Tip()
		blockHeight := msgBlock.Header.Height
		block := fnoutil.NewBlock(msgBlock)
		t.Logf("Testing block template %s (hash %s, height %d)",
			g.TipName(), block.Hash(), blockHeight)

		err := chain.CheckConnectBlockTemplate(block)
		if err == nil {
			t.Fatalf("block template %q (hash %s, height %d) should "+
				"not have been accepted", g.TipName(), block.Hash(),
				blockHeight)
		}

		// Ensure the error code is of the expected type and the reject
		// code matches the value specified in the test instance.
		rerr, ok := err.(RuleError)
		if !ok {
			t.Fatalf("block template %q (hash %s, height %d) "+
				"returned unexpected error type -- got %T, want "+
				"blockchain.RuleError", g.TipName(),
				block.Hash(), blockHeight, err)
		}
		if rerr.ErrorCode != code {
			t.Fatalf("block template %q (hash %s, height %d) does "+
				"not have expected reject code -- got %v, want %v",
				g.TipName(), block.Hash(), blockHeight,
				rerr.ErrorCode, code)
		}
	}

	// changeNonce is a munger that modifies the block by changing the header
	// nonce to a pseudo-random value.
	prng := mrand.New(mrand.NewSource(0))
	changeNonce := func(b *wire.MsgBlock) {
		// Change the nonce so the block isn't actively solved.
		b.Header.Nonce = prng.Uint32()
	}

	// Shorter versions of useful params for convenience.
	ticketsPerBlock := params.TicketsPerBlock
	coinbaseMaturity := params.CoinbaseMaturity
	stakeEnabledHeight := params.StakeEnabledHeight
	stakeValidationHeight := params.StakeValidationHeight

	// ---------------------------------------------------------------------
	// Premine block templates.
	// ---------------------------------------------------------------------

	// Produce a premine block with too much coinbase and ensure the block
	// template is rejected.
	//
	//   genesis
	//          \-> bpbad
	g.CreatePremineBlock("bpbad", 1)
	g.AssertTipHeight(1)
	rejectedBlockTemplate(ErrBadCoinbaseValue)

	// Produce a valid, but unsolved premine block and ensure the block template
	// is accepted while the unsolved block is rejected.
	//
	//   genesis
	//          \-> bpunsolved
	g.SetTip("genesis")
	bpunsolved := g.CreatePremineBlock("bpunsolved", 0, changeNonce)
	// Since the difficulty is so low in the tests, the block might still
	// end up being inadvertently solved.  It can't be checked inside the
	// munger because the block is finalized after the function returns and
	// those changes could also inadvertently solve the block.  Thus, just
	// increment the nonce until it's not solved and then replace it in the
	// generator's state.
	{
		origHash := bpunsolved.BlockHash()
		for chaingen.IsSolved(&bpunsolved.Header) {
			bpunsolved.Header.Nonce++
		}
		g.UpdateBlockState("bpunsolved", origHash, "bpunsolved", bpunsolved)
	}
	g.AssertTipHeight(1)
	acceptedBlockTemplate()
	rejected(ErrHighHash)
	expectTip("genesis")

	// Produce a valid and solved premine block.
	//
	//   genesis -> bp
	g.SetTip("genesis")
	g.CreatePremineBlock("bp", 0)
	g.AssertTipHeight(1)
	accepted()

	// ---------------------------------------------------------------------
	// Generate enough blocks to have mature coinbase outputs to work with.
	//
	// Also, ensure that each block is considered a valid template along the
	// way.
	//
	//   genesis -> bp -> bm0 -> bm1 -> ... -> bm#
	// ---------------------------------------------------------------------

	var tipName string
	for i := uint16(0); i < coinbaseMaturity; i++ {
		blockName := fmt.Sprintf("bm%d", i)
		g.NextBlock(blockName, nil, nil)
		g.SaveTipCoinbaseOuts()
		acceptedBlockTemplate()
		accepted()
		tipName = blockName
	}
	g.AssertTipHeight(uint32(coinbaseMaturity) + 1)

	// ---------------------------------------------------------------------
	// Generate block templates that include invalid ticket purchases.
	// ---------------------------------------------------------------------

	// Create a block template with a ticket that claims too much input
	// amount.
	//
	//   ... -> bm#
	//             \-> btixt1
	tempOuts := g.OldestCoinbaseOuts()
	tempTicketOuts := tempOuts[1:]
	g.NextBlock("btixt1", nil, tempTicketOuts, func(b *wire.MsgBlock) {
		changeNonce(b)
		b.STransactions[3].TxIn[0].ValueIn--
	})
	rejectedBlockTemplate(ErrFraudAmountIn)

	// Create a block template with a ticket that does not pay enough.
	//
	//   ... -> bm#
	//             \-> btixt2
	g.SetTip(tipName)
	g.NextBlock("btixt2", nil, tempTicketOuts, func(b *wire.MsgBlock) {
		changeNonce(b)
		b.STransactions[2].TxOut[0].Value--
	})
	rejectedBlockTemplate(ErrNotEnoughStake)

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake enabled height while
	// creating ticket purchases that spend from the coinbases matured
	// above.  This will also populate the pool of immature tickets.
	//
	// Also, ensure that each block is considered a valid template along the
	// way.
	//
	//   ... -> bm# ... -> bse0 -> bse1 -> ... -> bse#
	// ---------------------------------------------------------------------

	// Use the already popped outputs.
	g.SetTip(tipName)
	g.NextBlock("bse0", nil, tempTicketOuts)
	g.SaveTipCoinbaseOuts()
	acceptedBlockTemplate()
	accepted()

	var ticketsPurchased int
	for i := int64(1); int64(g.Tip().Header.Height) < stakeEnabledHeight; i++ {
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		ticketsPurchased += len(ticketOuts)
		blockName := fmt.Sprintf("bse%d", i)
		g.NextBlock(blockName, nil, ticketOuts)
		g.SaveTipCoinbaseOuts()
		acceptedBlockTemplate()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeEnabledHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to reach the stake validation height while
	// continuing to purchase tickets using the coinbases matured above and
	// allowing the immature tickets to mature and thus become live.
	//
	// Also, ensure that each block is considered a valid template along the
	// way.
	//
	//   ... -> bse# -> bsv0 -> bsv1 -> ... -> bsv#
	// ---------------------------------------------------------------------

	targetPoolSize := g.Params().TicketPoolSize * ticketsPerBlock
	for i := int64(0); int64(g.Tip().Header.Height) < stakeValidationHeight; i++ {
		// Only purchase tickets until the target ticket pool size is
		// reached.
		outs := g.OldestCoinbaseOuts()
		ticketOuts := outs[1:]
		if ticketsPurchased+len(ticketOuts) > int(targetPoolSize) {
			ticketsNeeded := int(targetPoolSize) - ticketsPurchased
			if ticketsNeeded > 0 {
				ticketOuts = ticketOuts[1 : ticketsNeeded+1]
			} else {
				ticketOuts = nil
			}
		}
		ticketsPurchased += len(ticketOuts)

		blockName := fmt.Sprintf("bsv%d", i)
		g.NextBlock(blockName, nil, ticketOuts)
		g.SaveTipCoinbaseOuts()
		acceptedBlockTemplate()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight))

	// ---------------------------------------------------------------------
	// Generate enough blocks to have a known distance to the first mature
	// coinbase outputs for all tests that follow.  These blocks continue
	// to purchase tickets to avoid running out of votes.
	//
	// Also, ensure that each block is considered a valid template along the
	// way.
	//
	//   ... -> bsv# -> bbm0 -> bbm1 -> ... -> bbm#
	// ---------------------------------------------------------------------

	for i := uint16(0); i < coinbaseMaturity; i++ {
		outs := g.OldestCoinbaseOuts()
		blockName := fmt.Sprintf("bbm%d", i)
		g.NextBlock(blockName, nil, outs[1:])
		g.SaveTipCoinbaseOuts()
		acceptedBlockTemplate()
		accepted()
	}
	g.AssertTipHeight(uint32(stakeValidationHeight) + uint32(coinbaseMaturity))

	// Collect spendable outputs into two different slices.  The outs slice
	// is intended to be used for regular transactions that spend from the
	// output, while the ticketOuts slice is intended to be used for stake
	// ticket purchases.
	var outs []*chaingen.SpendableOut
	var ticketOuts [][]chaingen.SpendableOut
	for i := uint16(0); i < coinbaseMaturity; i++ {
		coinbaseOuts := g.OldestCoinbaseOuts()
		outs = append(outs, &coinbaseOuts[0])
		ticketOuts = append(ticketOuts, coinbaseOuts[1:])
	}

	// ---------------------------------------------------------------------
	// Generate block templates that build on ancestors of the tip.
	// ---------------------------------------------------------------------

	// Start by building a few of blocks at current tip (value in parens
	// is which output is spent):
	//
	//   ... -> b1(0) -> b2(1) -> b3(2)
	g.NextBlock("b1", outs[0], ticketOuts[0])
	accepted()

	g.NextBlock("b2", outs[1], ticketOuts[1])
	accepted()

	g.NextBlock("b3", outs[2], ticketOuts[2])
	accepted()

	// Create a block template that forks from b1.  It should not be allowed
	// since it is not the current tip or its parent.
	//
	//   ... -> b1(0) -> b2(1)  -> b3(2)
	//               \-> b2at(1)
	g.SetTip("b1")
	g.NextBlock("b2at", outs[1], ticketOuts[1], changeNonce)
	rejectedBlockTemplate(ErrInvalidTemplateParent)

	// Create a block template that forks from b2.  It should be accepted
	// because it is the current tip's parent.
	//
	//   ... -> b2(1) -> b3(2)
	//               \-> b3at(2)
	g.SetTip("b2")
	g.NextBlock("b3at", outs[2], ticketOuts[2], changeNonce)
	acceptedBlockTemplate()

	// ---------------------------------------------------------------------
	// Generate block templates that build on the tip's parent, but include
	// invalid votes.
	// ---------------------------------------------------------------------

	// Create a block template that forks from b2 (the tip's parent) with
	// votes that spend invalid tickets.
	//
	//   ... -> b2(1) -> b3(2)
	//               \-> b3bt(2)
	g.SetTip("b2")
	g.NextBlock("b3bt", outs[2], ticketOuts[1], changeNonce)
	rejectedBlockTemplate(ErrMissingTxOut)

	// Same as before but based on the current tip.
	//
	//   ... -> b2(1) -> b3(2)
	//                        \-> b4at(3)
	g.SetTip("b3")
	g.NextBlock("b4at", outs[3], ticketOuts[2], changeNonce)
	rejectedBlockTemplate(ErrMissingTxOut)

	// Create a block template that forks from b2 (the tip's parent) with
	// a vote that pays too much.
	//
	//   ... -> b2(1) -> b3(2)
	//               \-> b3ct(2)
	g.SetTip("b2")
	g.NextBlock("b3ct", outs[2], ticketOuts[2], func(b *wire.MsgBlock) {
		changeNonce(b)
		b.STransactions[0].TxOut[0].Value++
	})
	rejectedBlockTemplate(ErrSpendTooHigh)

	// Same as before but based on the current tip.
	//
	//   ... -> b2(1) -> b3(2)
	//                        \-> b4bt(3)
	g.SetTip("b3")
	g.NextBlock("b4bt", outs[3], ticketOuts[3], func(b *wire.MsgBlock) {
		changeNonce(b)
		b.STransactions[0].TxOut[0].Value++
	})
	rejectedBlockTemplate(ErrSpendTooHigh)

	// ---------------------------------------------------------------------
	// Generate block templates that build on the tip and its parent after a
	// forced reorg.
	// ---------------------------------------------------------------------

	// Create a fork from b2.  There should not be a reorg since b3 was seen
	// first.
	//
	//   ... -> b2(1) -> b3(2)
	//               \-> b3a(2)
	g.SetTip("b2")
	g.NextBlock("b3a", outs[2], ticketOuts[2])
	acceptedToSideChainWithExpectedTip("b3")

	// Force tip reorganization to b3a.
	//
	//   ... -> b2(1) -> b3a(2)
	//               \-> b3(2)
	forceTipReorg("b3", "b3a")
	expectTip("b3a")

	// Create a block template that forks from b2 (the tip's parent) and
	// ensure it is still accepted after the forced reorg.
	//
	//   ... -> b2(1) -> b3a(2)
	//               \-> b3dt(2)
	g.SetTip("b2")
	g.NextBlock("b3dt", outs[2], ticketOuts[2], changeNonce)
	acceptedBlockTemplate()
	expectTip("b3a") // Ensure chain tip didn't change.

	// Create a block template that builds on the current tip and ensure it
	// it is still accepted after the forced reorg.
	//
	//   ... -> b2(1) -> b3a(2)
	//                         \-> b4ct(3)
	g.SetTip("b3a")
	g.NextBlock("b4ct", outs[3], ticketOuts[3], changeNonce)
	acceptedBlockTemplate()
}
