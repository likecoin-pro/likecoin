package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
	"github.com/likecoin-pro/likecoin/crypto/patricia"
)

type Block struct {
	*BlockHeader
	Txs []*Transaction `json:"txs"`
}

func NewBlock(h *BlockHeader, txs []*Transaction) *Block {
	return &Block{h, txs}
}

type BCContext interface {
	State() *state.State
	StateTree() *patricia.Tree
	ChainTree() *patricia.Tree
}

// todo: StateTree() move to *State
// todo: tx.Execute() -> stateUpdates, stateRoot, err

func GenerateNewBlock(
	pre *BlockHeader,
	txs []*Transaction,
	prv *crypto.PrivateKey,
	bc BCContext,
) (block *Block, err error) {
	return GenerateNewBlockEx(pre, txs, prv, bc, timestamp(), 0)
}

func GenerateNewBlockEx(
	pre *BlockHeader,
	txs []*Transaction,
	prv *crypto.PrivateKey,
	bc BCContext,
	timestamp int64,
	nonce uint64,
) (block *Block, err error) {

	block = &Block{&BlockHeader{
		Version:   0,
		Network:   config.NetworkID,
		ChainID:   pre.ChainID,
		Num:       pre.Num + 1,
		PrevHash:  pre.Hash(),
		Timestamp: timestamp,
		Nonce:     nonce,
		Miner:     prv.PublicKey,
	}, txs}

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
	block.TxRoot = block.txRoot()
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

// Size returns block-header size + txs size
func (b *Block) Size() int {
	return len(b.Encode())
}

func (b *Block) CountTxs() int {
	return len(b.Txs)
}

func (b *Block) Encode() []byte {
	return bin.Encode(b.BlockHeader, b.Txs)
}

func (b *Block) Decode(data []byte) (err error) {
	return bin.Decode(data, &b.BlockHeader, &b.Txs)
}

func (b *Block) Verify(pre *BlockHeader) error {
	// verify block header
	if err := b.BlockHeader.VerifyHeader(pre); err != nil {
		return err
	}
	// verify block txs
	if err := b.verifyTxs(); err != nil {
		return err
	}
	return nil
}

func (b *Block) verifyTxs() error {
	if len(b.Txs) == 0 {
		return ErrEmptyBlock
	}
	for _, tx := range b.Txs {
		// check tx-chain info
		if tx.ChainID != b.ChainID {
			return ErrInvalidChainID
		}
		if tx.Network != b.Network {
			return ErrTxInvalidNetworkID
		}
	}
	if txRoot := b.txRoot(); !bytes.Equal(b.TxRoot, txRoot) {
		return ErrInvalidTxsMerkleRoot
	}
	return nil
}

func (b *Block) txRoot() []byte {
	//tree:=patricia.NewTree(nil)
	//for idx, it := range txs {
	//	tree.Put(bin.Encode(idx), it.TxStHash())
	//}
	//root,_:= tree.Root()
	//return root

	var hh [][]byte
	for _, it := range b.Txs {
		hh = append(hh, it.TxStHash())
	}
	return merkle.Root(hh...)
}
