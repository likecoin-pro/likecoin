package patricia

import (
	"bytes"
	"errors"
)

type Storage interface {
	Get(key []byte) ([]byte, error)
	Put(key, data []byte) error
	//Delete(key []byte) error
}

type Tree struct {
	db     Storage
	keyPfx []byte
	root   []byte
}

var (
	errKeyNotFound     = errors.New("patricia: key not found tree-node data")
	errInvalidNodeData = errors.New("patricia: invalid tree-node data")
)

func NewTree(db Storage, keyPfx []byte) *Tree {
	if db == nil {
		db = MemoryStorage{}
	}
	return &Tree{db: db, keyPfx: keyPfx}
}

func (t *Tree) Root() (root []byte, err error) {
	if t.root != nil {
		return t.root, nil
	}
	if nd, err := t.getNode(nil); err != nil {
		return nil, err
	} else if nd != nil {
		t.root = nd.hash()
	}
	return t.root, nil
}

func (t *Tree) Put(key []byte) (err error) {
	t.root, err = t.put(make([]byte, 0, 10), key)
	return
}

func (t *Tree) GetProof(key []byte) ([]byte, error) {
	return t.proof(make([]byte, 0, 10), key)
}

//--------------------------------------------------
func (t *Tree) getNode(key []byte) (nd *node, err error) {
	data, err := t.db.Get(t.dbKey(key))
	if err != nil {
		return
	}
	if len(data) == 0 {
		return nil, nil
	}
	nd = new(node)
	_, err = nd.decode(data)
	return
}

func idx(key []byte, lv int) uint8 {
	if lv%2 == 0 {
		return key[lv/2] >> 4
	} else {
		return key[lv/2] & 0x0f
	}
}

func (t *Tree) dbKey(key []byte) []byte {
	if len(t.keyPfx) == 0 {
		return key
	}
	return append(t.keyPfx, key...)
}

func (t *Tree) put(path, key []byte) (newHash []byte, err error) {
	nd, err := t.getNode(path)
	if err != nil {
		return
	} else if nd == nil {
		nd = &node{val: key}
	} else {
		lv := len(path)
		if len(nd.val) != 0 {
			val := nd.val
			if bytes.Equal(val, key) {
				return nd.hash(), nil
			}
			i := idx(val, lv)
			nd.val = nil
			nd.hashes[i], err = t.put(append(path, i), val)
			if err != nil {
				return
			}
		}
		i := idx(key, lv)
		nd.hashes[i], err = t.put(append(path, i), key)
		if err != nil {
			return
		}
	}
	err = t.db.Put(t.dbKey(path), nd.encode())
	return nd.hash(), err
}

func (t *Tree) proof(path, key []byte) (proof []byte, err error) {
	nd, err := t.getNode(path)
	if err != nil {
		return
	}
	if nd == nil { // null leaf
		return nil, errKeyNotFound
	}
	if nd.val != nil { // leaf
		if !bytes.Equal(nd.val, key) {
			// todo: return proof of empty
			return nil, errKeyNotFound
		}
		return nil, nil
	}
	i := idx(key, len(path))
	proof, err = t.proof(append(path, i), key)
	proof = append(proof, nd.proof(int(i))...)
	return
}
