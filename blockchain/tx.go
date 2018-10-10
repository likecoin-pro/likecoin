package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/denisskin/bin"
	"github.com/likecoin-pro/likecoin/blockchain/state"
	"github.com/likecoin-pro/likecoin/commons/enc"
	"github.com/likecoin-pro/likecoin/commons/hex"
	"github.com/likecoin-pro/likecoin/config"
	"github.com/likecoin-pro/likecoin/crypto"
	"github.com/likecoin-pro/likecoin/crypto/merkle"
)

type TxType = uint8

type Transaction struct {

	// Tx data
	Type      TxType            // tx-type
	Version   int               // tx version
	Network   int               // networkID
	ChainID   uint64            //
	Nonce     uint64            // sender nonce (by default: Unix-time in µsec)
	Data      []byte            // encoded tx-object
	Reserved1 []byte            //
	Reserved2 []byte            //
	Sender    *crypto.PublicKey // tx-sender
	Sig       []byte            // tx-sender signature

	// Chain data
	StateUpdates state.Values // state changes (is not filled by sender)

	// not imported fields
	blockNum uint64    //
	blockIdx int       //
	blockTs  int64     //
	_obj     TxObject  //
	bc       BCContext //
}

func NewTx(cfg *Config, sender *crypto.PrivateKey, nonce uint64, obj TxObject) *Transaction {
	if nonce == 0 {
		nonce = uint64(timestamp())
	}
	tx := &Transaction{
		Type:    typeByObject(obj), //
		Version: 0,                 //
		Network: cfg.NetworkID,     //
		ChainID: cfg.ChainID,       //
		Sender:  sender.PublicKey,  //
		Nonce:   nonce,             //
		Data:    obj.Encode(),      // encoded tx-object
	}
	obj.SetContext(tx)
	tx.Sig = sender.Sign(tx.Hash())
	return tx
}

var (
	ErrTxEmptySender      = errors.New("tx-verify-error: empty tx-sender")
	ErrTxEmptyData        = errors.New("tx-verify-error: empty tx-data")
	ErrTxInvalidData      = errors.New("tx-verify-error: invalid tx-data")
	ErrTxInvalidChainID   = errors.New("tx-verify-error: invalid chain-id")
	ErrTxInvalidNetworkID = errors.New("tx-verify-error: invalid network-id")
	ErrTxDataIsTooLong    = errors.New("tx-verify-error: tx is too long")
)

func (tx *Transaction) String() string {
	if obj, err := tx.Object(); err == nil {
		return enc.IndentJSON(obj)
	}
	return enc.IndentJSON(tx)
}

func (tx *Transaction) ID() uint64 {
	return TxIDByHash(tx.Hash())
}

func (tx *Transaction) StrID() string {
	return hex.EncodeUint(tx.ID())
}

func (tx *Transaction) SenderAddress() crypto.Address {
	if tx != nil && tx.Sender != nil {
		return tx.Sender.Address()
	}
	return crypto.NilAddress
}

func (tx *Transaction) SenderNick() (nick string, err error) {
	if tx != nil && tx.bc != nil && tx.Sender != nil {
		nick, err = tx.bc.UsernameByID(tx.Sender.Address().ID())
	}
	return
}

// Hash returns hash of senders data
func (tx *Transaction) Hash() []byte {
	return crypto.Hash256(
		tx.Type,
		tx.Version,
		tx.Network,
		tx.ChainID,
		tx.Nonce,
		tx.Data,
		tx.Reserved1,
		tx.Reserved2,
		tx.Sender,
	)
}

func (tx *Transaction) TxStHash() []byte {
	return merkle.Root(tx.Hash(), tx.StateUpdates.Hash())
}

func (tx *Transaction) Size() int {
	return len(tx.Encode())
}

func (tx *Transaction) StrType() string {
	return TxTypeStr(tx.Type)
}

func (tx *Transaction) Equal(tx1 *Transaction) bool {
	return bytes.Equal(tx.Encode(), tx1.Encode())
}

func (tx *Transaction) TxObject() TxObject {
	obj, _ := tx.Object()
	return obj
}

func (tx *Transaction) SetBlockInfo(bc BCContext, blockNum uint64, blockTxIdx int, blockTs int64) {
	tx.bc, tx.blockNum, tx.blockIdx, tx.blockTs = bc, blockNum, blockTxIdx, blockTs
}

func (tx *Transaction) BCContext() BCContext {
	if tx != nil {
		return tx.bc
	}
	return nil
}

func (tx *Transaction) BlockNum() uint64 {
	return tx.blockNum
}

func (tx *Transaction) BlockIdx() int {
	return tx.blockIdx
}

func (tx *Transaction) BlockTs() int64 {
	return tx.blockTs
}

func (tx *Transaction) Seq() uint64 {
	return (tx.blockNum << 32) | uint64(tx.blockIdx)
}

func (tx *Transaction) Encode() []byte {
	if len(tx.Data) == 0 {
		panic(ErrTxEmptyData)
	}
	return bin.Encode(
		tx.Type,
		tx.Version,
		tx.Network,
		tx.ChainID,
		tx.Nonce,
		tx.Data,
		tx.Reserved1,
		tx.Reserved2,
		tx.Sender,
		tx.Sig,
		tx.StateUpdates,
	)
}

func (tx *Transaction) Decode(data []byte) (err error) {
	return bin.Decode(data,
		&tx.Type,
		&tx.Version,
		&tx.Network,
		&tx.ChainID,
		&tx.Nonce,
		&tx.Data,
		&tx.Reserved1,
		&tx.Reserved2,
		&tx.Sender,
		&tx.Sig,
		&tx.StateUpdates,
	)
}

func (tx *Transaction) Object() (obj TxObject, err error) {
	if tx._obj != nil {
		return tx._obj, nil
	}
	obj, err = newObjectByType(tx.Type)
	if err != nil {
		return
	}
	obj.SetContext(tx)
	if err = obj.Decode(tx.Data); err != nil {
		return
	}
	tx._obj = obj
	return
}

// Timestamp returns user timestamp from nonce
func (tx *Transaction) Timestamp() time.Time {
	return time.Unix(0, int64(tx.blockTs)*1e3)
}

func (tx *Transaction) Verify(cfg *Config) error {

	//-- verify transaction data
	if tx.Network != cfg.NetworkID {
		return ErrTxInvalidNetworkID
	}
	if tx.ChainID != cfg.ChainID {
		return ErrTxInvalidChainID
	}
	if len(tx.Data) == 0 {
		return ErrTxEmptyData
	}
	if tx.Type != 0 && len(tx.Data) > config.MaxTxDataSize {
		return ErrTxDataIsTooLong
	}
	if tx.Sender == nil || tx.Sender.Empty() {
		return ErrTxEmptySender
	}
	txObj, err := tx.Object()
	if err != nil {
		return err
	}
	if err := txObj.Verify(); err != nil {
		return err
	}

	//-- verify sender signature
	if !tx.Sender.Verify(tx.Hash(), tx.Sig) {
		return ErrInvalidBlockSig
	}
	return nil
}

// Execute executes tx, changes state, returns state-updates
func (tx *Transaction) Execute(s *state.State) (updates state.Values, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("tx.Execute-panic: %v", r)
		}
	}()

	obj, err := tx.Object()
	if err != nil {
		return
	}

	newState := s.NewSubState()

	obj.Execute(newState)

	updates = newState.Values()

	return
}

func TxIDByHash(txHash []byte) uint64 {
	return bin.BytesToUint64(txHash[:8])
}
