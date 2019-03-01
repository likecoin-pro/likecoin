package blockchain

import (
	"errors"
	"flag"
	"sync/atomic"
	"time"
)

var (
	ErrEmptyBlock           = errors.New("block.Verify-error: empty block")
	ErrInvalidGenesisBlock  = errors.New("block.Verify-error: invalid genesis block")
	ErrEmptyMinerKey        = errors.New("block.Verify-error: empty miner public key")
	ErrInvalidMinerKey      = errors.New("block.Verify-error: invalid miner public key")
	ErrInvalidBlockSig      = errors.New("block.Verify-error: invalid signature")
	ErrInvalidBlockNum      = errors.New("block.Verify-error: invalid block num")
	ErrInvalidBlockTs       = errors.New("block.Verify-error: invalid block timestamp")
	ErrInvalidNetwork       = errors.New("block.Verify-error: invalid network ID")
	ErrInvalidChainID       = errors.New("block.Verify-error: invalid chain ID")
	ErrInvalidPrevHash      = errors.New("block.Verify-error: invalid previous block hash")
	ErrInvalidTxsMerkleRoot = errors.New("block.Verify-error: invalid txs merkle root")
)

var (
	testMode  = flag.Lookup("test.v") != nil
	testTimer int64
)

// Timestamp returns current timestamp in µsec
func Timestamp() int64 {
	if testMode {
		return atomic.AddInt64(&testTimer, 1)
	}
	return TimeToInt(time.Now())
}

func TimeToInt(t time.Time) int64 {
	return t.UnixNano() / 1e3
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
