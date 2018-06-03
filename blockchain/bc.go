package blockchain

import (
	"errors"
	"flag"
	"sync/atomic"
	"time"
)

var (
	ErrEmptyBlock          = errors.New("block.Verify-error: Empty block")
	ErrInvalidGenesisBlock = errors.New("block.Verify-error: Invalid genesis block")
	ErrEmptyNodeKey        = errors.New("block.Verify-error: Empty node public key")
	ErrInvalidNodeKey      = errors.New("block.Verify-error: Invalid node public key")
	ErrInvalidSign         = errors.New("block.Verify-error: Invalid sign")
	ErrInvalidNum          = errors.New("block.Verify-error: Invalid block num")
	ErrInvalidChainID      = errors.New("block.Verify-error: Invalid chain ID")
	ErrInvalidPrevHash     = errors.New("block.Verify-error: Invalid previous hash")
	ErrInvalidMerkleRoot   = errors.New("block.Verify-error: Invalid Merkle Root")
)

var (
	testMode  = flag.Lookup("test.v") != nil
	testTimer int64
)

// returns current timestamp in microsec
func timestamp() int64 {
	if testMode {
		return atomic.AddInt64(&testTimer, 1)
	}
	return time.Now().UnixNano() / 1e3
}

func init() {
	time.Local = time.UTC

	if testMode {
		InitTestTimer()
	}
}

func InitTestTimer() {
	if !testMode {
		panic("InitTestTimer can be call in test mode only")
	}
	atomic.StoreInt64(&testTimer, 1.5e15)
}
