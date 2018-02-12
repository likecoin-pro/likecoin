package blockchain

import (
	"errors"
	"time"
)

var (
	ErrEmptyBlock          = errors.New("block.Verify-error: Empty block")
	ErrInvalidGenesisBlock = errors.New("block.Verify-error: Invalid genesis block")
	ErrEmptyNodeKey        = errors.New("block.Verify-error: Empty node public key")
	ErrInvalidNodeKey      = errors.New("block.Verify-error: Invalid node public key")
	ErrInvalidSign         = errors.New("block.Verify-error: Invalid sign")
	ErrInvalidNum          = errors.New("block.Verify-error: Invalid block num")
	ErrInvalidPrevHash     = errors.New("block.Verify-error: Invalid previous hash")
	ErrInvalidMerkleRoot   = errors.New("block.Verify-error: Invalid Merkle Root")
)

// returns current timestamp in microsec
var timestamp = func() int64 {
	return time.Now().UnixNano() / 1e3
}

func init() {
	time.Local = time.UTC
}
