// Copyright (c) 2013-2016 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.
package blockchain_test

import (
	"bytes"
	"compress/bzip2"
	"encoding/gob"
	"os"
	"path/filepath"
	"testing"

	bc "github.com/HcashOrg/hcashd/blockchain"
	"github.com/HcashOrg/hcashutil"
)

// TestBlockchainFunction tests the various blockchain API to ensure proper
// functionality.
// DOESN'T WORK YET
func DNWTestBlockchainFunctions(t *testing.T) {
	// Create a new database and chain instance to run tests against.
	//chain, teardownFunc, err := blockchain.chainSetup("validateunittests", SimNetParams)
	chain, teardownFunc, err := bc.SetupTestChain("validateunittests", bc.SimNetParams)
	if err != nil {
		t.Errorf("Failed to setup chain instance: %v", err)
		return
	}
	defer teardownFunc()

	// The genesis block should fail to connect since it's already inserted.
	genesisBlock := bc.SimNetParams.GenesisBlock
	err = chain.CheckConnectBlock(hcashutil.NewBlock(genesisBlock), false)
	if err == nil {
		t.Errorf("CheckConnectBlock: Did not receive expected error")
	}

	// Load up the rest of the blocks up to HEAD~1.
	filename := filepath.Join("testdata/", "hcash0to168.tar.bz2")
	fi, err := os.Open(filename)
	if err != nil {
		t.Errorf("Unable to open %s: %v", filename, err)
	}
	bcStream := bzip2.NewReader(fi)
	defer fi.Close()

	// Create a buffer of the read file.
	bcBuf := new(bytes.Buffer)
	bcBuf.ReadFrom(bcStream)

	// Create decoder from the buffer and a map to store the data.
	bcDecoder := gob.NewDecoder(bcBuf)
	blockChain := make(map[int64][]byte)

	// Decode the blockchain into the map.
	if err := bcDecoder.Decode(&blockChain); err != nil {
		t.Errorf("error decoding test blockchain: %v", err.Error())
	}

	// Insert blocks 1 to 168 and perform various tests.
	for i := 1; i <= 168; i++ {
		bl, err := hcashutil.NewBlockFromBytes(blockChain[int64(i)])
		if err != nil {
			t.Errorf("NewBlockFromBytes error: %v", err.Error())
		}

		//_, _, err = chain.ProcessBlock(bl, blockchain.BFNone)
		_, _, err = chain.ProcessBlock(bl, bc.BFNone)
		if err != nil {
			t.Fatalf("ProcessBlock error at height %v: %v", i, err.Error())
		}
	}

	val, err := chain.TicketPoolValue()
	if err != nil {
		t.Errorf("Failed to get ticket pool value: %v", err)
	}
	expectedVal := hcashutil.Amount(3495091704)
	if val != expectedVal {
		t.Errorf("Failed to get correct result for ticket pool value; "+
			"want %v, got %v", expectedVal, val)
	}

	a, _ := hcashutil.DecodeAddress("SsbKpMkPnadDcZFFZqRPY8nvdFagrktKuzB")
	hs, err := chain.TicketsWithAddress(a)
	if err != nil {
		t.Errorf("Failed to do TicketsWithAddress: %v", err)
	}
	expectedLen := 223
	if len(hs) != expectedLen {
		t.Errorf("Failed to get correct number of tickets for "+
			"TicketsWithAddress; want %v, got %v", expectedLen, len(hs))
	}

	totalSubsidy := chain.TotalSubsidy()
	expectedSubsidy := int64(35783267326630)
	if expectedSubsidy != totalSubsidy {
		t.Errorf("Failed to get correct total subsidy for "+
			"TotalSubsidy; want %v, got %v", expectedSubsidy,
			totalSubsidy)
	}
}
