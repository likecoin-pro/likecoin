package config

import (
	"flag"
	"fmt"
	"os"
)

const (
	Version         = "0.1"
	ApplicationName = "likecoin-core"
)

// program arguments
var (
	// network params
	NetworkID = 0         // 0-work, 1-test
	ChainID   = uint64(1) // blockchain ID

	// program options
	VerifyTransactions = false                              // by default verify only block-headers
	DataDir            = os.Getenv("HOME") + "/Likecoin.db" //
	DebugMode          = false                              //
)

func ParseArgs() {
	var (
		argHelp    = flag.Bool("help", false, "Show this help")
		argVersion = flag.Bool("version", false, "Show software version")
	)
	flag.IntVar(&NetworkID, "network-id", NetworkID, "NetworkID")
	flag.Uint64Var(&ChainID, "chain-id", ChainID, "ChainID")
	flag.BoolVar(&VerifyTransactions, "verify-tx", VerifyTransactions, "Verify each transactions")
	flag.StringVar(&DataDir, "db", DataDir, "Database dir")
	//flag.BoolVar(&AutoUpdate, "autoupdate", AutoUpdate, "Autoupdate soft")
	flag.BoolVar(&DebugMode, "debug", DebugMode, "Debug mode")
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
