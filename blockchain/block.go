package blockchain

import (
	"bytes"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
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
	TransactionByID(txID uint64) (*Transaction, error)
	UsernameByID(userID uint64) (nick string, err error)
}

// todo: StateTree() move to *State

func GenerateNewBlock(
	pre *BlockHeader,
	txs []*Transaction,
	prv *crypto.PrivateKey,
	bc BCContext,
	nonce uint64,
) (block *Block, err error) {

	st := bc.State()
	validTxs := txs[:0]
	for _, tx := range txs {
		if tx, err := bc.TransactionByID(tx.ID()); err != nil {
			return nil, err
		} else if tx != nil {
			continue // skip
		}
		if upd, err := tx.Execute(st); err == nil {
			tx.StateUpdates = upd
			st.Apply(upd)
			validTxs = append(validTxs, tx)
		}
	}
	if len(validTxs) == 0 {
		return nil, nil
	}

	block = &Block{&BlockHeader{
		Version:   0,
		Network:   pre.Network,
		ChainID:   pre.ChainID,
		Num:       pre.Num + 1,
		PrevHash:  pre.Hash(),
		Timestamp: timestamp(),
		Nonce:     nonce,
		Miner:     prv.PublicKey,
	}, validTxs}

	stTree := bc.StateTree()
	for _, tx := range block.Txs {
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
func (b *Block) Size() int64 {
	return int64(len(b.Encode()))
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

func (b *Block) Verify(pre *BlockHeader, bcCfg *Config) error {
	// verify block header
	if err := b.BlockHeader.VerifyHeader(pre, bcCfg); err != nil {
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
