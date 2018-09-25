package blockchain

import (
	"flag"
	"os"
)

type Config struct {
	NetworkID      int    // 0-work, 1-test
	ChainID        uint64 // blockchain ID
	DataDir        string //
	VerifyTxsLevel int    // by default verify only block-headers
	VacuumDB       bool   //
}

const (
	VerifyTxLevel0 = 0 // verify only block headers
	VerifyTxLevel1 = 1 // verify each tx in block

	NetworkWorking = 0 //
	NetworkTest    = 1 //
)

func NewConfig() *Config {
	cfg := &Config{
		NetworkID:      NetworkWorking,
		ChainID:        1,
		VerifyTxsLevel: VerifyTxLevel1,
		VacuumDB:       false,
		DataDir:        os.Getenv("HOME") + "/likecd.db",
	}
	flag.IntVar(&cfg.NetworkID, "network-id", cfg.NetworkID, "Network ID (0 - work network; 1 - test network")
	flag.Uint64Var(&cfg.ChainID, "chain-id", cfg.ChainID, "Chain ID")
	flag.IntVar(&cfg.VerifyTxsLevel, "verify-level", cfg.VerifyTxsLevel, "Verify tx level (0 - only block headers; 1 - each tx-state)")
	flag.BoolVar(&cfg.VacuumDB, "vacuum", cfg.VacuumDB, "Vacuum DB after start")
	flag.StringVar(&cfg.DataDir, "db", cfg.DataDir, "Database dir")
	return cfg
}
