package blockchain

import (
	"bytes"

	"fmt"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
)

type Block struct {
	Version   int       `json:"version"`       // version
	Network   int       `json:"network"`       // networkID
	ChainID   uint64    `json:"chain"`         //
	Num       uint64    `json:"height"`        // number of block in blockchain
	Timestamp int64     `json:"timestamp"`     // timestamp of block in µsec
	PrevHash  bin.Bytes `json:"previous_hash"` // hash of previous block
	TxRoot    bin.Bytes `json:"tx_root"`       // merkle root of block-transactions
	StateRoot bin.Bytes `json:"state_root"`    // patricia root of global state
	ChainRoot bin.Bytes `json:"chain_root"`    // patricia root of chain

	// miner params
	Nonce uint64            `json:"nonce"` //
	Miner *crypto.PublicKey `json:"miner"` // miner public-key
	Sig   bin.Bytes         `json:"sig"`   // miner signature  := minerKey.Sign( blockHash + chainRoot )

	// reserved
	Reserved1 []byte `json:"-"`
	Reserved2 []byte `json:"-"`
	Reserved3 []byte `json:"-"`
}

type BCContext interface {
	State() *state.State
	StateTree() *patricia.Tree
	ChainTree() *patricia.Tree
}

// todo: StateTree() move to *State
// todo: tx.Execute() -> stateUpdates, stateRoot, err

func NewBlock(
	pre *Block,
	txs []*Transaction,
	prv *crypto.PrivateKey,
	bc BCContext,
) (block *Block, err error) {

	block = &Block{
		Version:   0,
		Network:   config.NetworkID,
		ChainID:   pre.ChainID,
		Num:       pre.Num + 1,
		PrevHash:  pre.Hash(),
		Timestamp: timestamp(),
		Nonce:     0,
		Miner:     prv.PublicKey,
	}

	st := bc.State()
	stTree := bc.StateTree()
	for _, tx := range txs {
		tx.StateUpdates, err = tx.Execute(st)
		if err != nil {
			return
		}
		st.Apply(tx.StateUpdates)
		for _, v := range tx.StateUpdates {
			if v.ChainID == block.ChainID {
				stTree.Put(v.StateKey(), v.Balance.Bytes())
			}
		}
	}
	block.TxRoot = txRoot(txs)
	block.StateRoot, err = stTree.Root()
	if err != nil {
		return nil, err
	}

	chainTree := bc.ChainTree()
	err = chainTree.PutVar(block.Num, block.Hash())
	if err != nil {
		return nil, err
	}
	block.ChainRoot, err = chainTree.Root()
	if err != nil {
		return nil, err
	}

	// set signature( b.Hash + chainRoot )
	block.Sig = prv.Sign(block.sigHash())

	return
}

func (b *Block) String() string {
	h := b.Hash()
	return fmt.Sprintf("[BLOCK-%d 0x%x size:%d]", b.Num, h[:8], b.Size())
}

// block.Hash + chainRoot
func (b *Block) sigHash() []byte {
	return merkle.Root(b.Hash(), b.ChainRoot)
}

func (b *Block) Hash() []byte {
	return crypto.Hash256(
		b.Version,
		b.ChainID,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.TxRoot,
		b.StateRoot,
		b.Nonce,
		b.Miner,
		b.Reserved1,
		b.Reserved2,
		b.Reserved3,
	)
}

// Size returns block-header size
func (b *Block) Size() int {
	return len(b.Encode())
}

func (b *Block) Encode() []byte {
	return bin.Encode(
		b.Version,
		b.Network,
		b.ChainID,
		b.Num,
		b.Timestamp,
		b.PrevHash,
		b.TxRoot,
		b.StateRoot,
		b.ChainRoot,
		b.Nonce,
		b.Miner,
		b.Reserved1,
		b.Reserved2,
		b.Reserved3,
		b.Sig,
	)
}

func (b *Block) Decode(data []byte) (err error) {
	return bin.Decode(data,
		&b.Version,
		&b.Network,
		&b.ChainID,
		&b.Num,
		&b.Timestamp,
		&b.PrevHash,
		&b.TxRoot,
		&b.StateRoot,
		&b.ChainRoot,
		&b.Nonce,
		&b.Miner,
		&b.Reserved1,
		&b.Reserved2,
		&b.Reserved3,
		&b.Sig,
	)
}

func (b *Block) Verify(pre *Block) error {
	if b.Network != config.NetworkID {
		return ErrInvalidNetwork
	}
	if b.ChainID != config.ChainID {
		return ErrInvalidChainID
	}
	blockHash := b.Hash()
	if b.Num == 0 && bytes.Equal(blockHash, GenesisBlock().Hash()) { // is genesis
		return ErrInvalidGenesisBlock
	}
	if pre != nil {
		if b.Network != pre.Network {
			return ErrInvalidNetwork
		}
		if b.ChainID != pre.ChainID {
			return ErrInvalidChainID
		}
		if b.Num != pre.Num+1 {
			return ErrInvalidBlockNum
		}
		if !bytes.Equal(b.PrevHash, pre.Hash()) {
			return ErrInvalidPrevHash
		}
	}
	if b.Miner.Empty() {
		return ErrEmptyMinerKey
	}
	if !b.Miner.Equal(config.MasterPublicKey) {
		return ErrInvalidMinerKey
	}
	if !b.Miner.Verify(b.sigHash(), b.Sig) {
		return ErrInvalidBlockSig
	}
	return nil
}

func (b *Block) VerifyTxs(txs []*Transaction) error {

	if len(txs) == 0 {
		return ErrEmptyBlock
	}

	for _, tx := range txs {
		// check tx-chain info
		if tx.ChainID != b.ChainID {
			return ErrInvalidChainID
		}
		if tx.Network != b.Network {
			return ErrTxInvalidNetworkID
		}
	}

	if txRoot := txRoot(txs); !bytes.Equal(b.TxRoot, txRoot) {
		return ErrInvalidTxsMerkleRoot
	}

	return nil
}

func txRoot(txs []*Transaction) []byte {
	//tree:=patricia.NewTree(nil)
	//for idx, it := range txs {
	//	tree.Put(bin.Encode(idx), it.TxStHash())
	//}
	//root,_:= tree.Root()
	//return root

	var hh [][]byte
	for _, it := range txs {
		hh = append(hh, it.TxStHash())
	}
	return merkle.Root(hh...)
}
