package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version         = "0.1"
	ApplicationName = "likecd"
)

const (
	VerifyTxLevel0 = 0 // verify only block headers
	VerifyTxLevel1 = 1 // verify each tx in block
)

// program arguments
var (
	// blockchain params
	NetworkID     = 0                                  // 0-work, 1-test
	ChainID       = uint64(1)                          // blockchain ID
	DataDir       = os.Getenv("HOME") + "/Likecoin.db" //
	VerifyTxLevel = VerifyTxLevel1                     // by default verify each tx state in blocks
)

func ParseArgs() {
	var (
		argHelp    = flag.Bool("help", false, "Show this help")
		argVersion = flag.Bool("version", false, "Show software version")
	)
	flag.IntVar(&NetworkID, "network-id", NetworkID, "NetworkID (0 - work network; 1 - test network")
	flag.Uint64Var(&ChainID, "chain-id", ChainID, "ChainID")
	flag.IntVar(&VerifyTxLevel, "verify-level", VerifyTxLevel, "Verify tx level (0 - only block headers; 1 - each tx-state)")
	flag.StringVar(&DataDir, "db", DataDir, "Database dir")
	flag.Parse()

	if *argHelp {
		flag.PrintDefaults()
		os.Exit(0)
		return
	}
	if *argVersion {
		fmt.Printf("%s (%s) version %s\n", ApplicationName, ProgramType, Version)
		os.Exit(0)
		return
	}
}
