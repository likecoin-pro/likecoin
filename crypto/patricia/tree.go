package patricia

import (
	"bytes"
	"errors"

	"github.com/denisskin/bin"
)

type Tree struct {
	db   Storage
	root []byte
	puts map[string]*node
}

var (
	errKeyNotFound     = errors.New("patricia: key not found tree-node data")
	errInvalidNodeData = errors.New("patricia: invalid tree-node data")
)

func NewTree(db Storage) *Tree {
	if db == nil {
		db = NewMemoryStorage(nil)
	}
	return &Tree{db: db}
}

func NewSubTree(db Storage, keyPrefix []byte) *Tree {
	return NewTree(NewSubStorage(db, keyPrefix))
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

func (t *Tree) PutVar(key, value interface{}) (err error) {
	return t.Put(encode(key), encode(value))
}

func decode(data []byte, v interface{}) (err error) {
	switch v := v.(type) {
	case *[]byte:
		*v = data
	case *string:
		*v = string(data)
	default:
		err = bin.Decode(data, v)
	}
	return
}

func encode(v interface{}) []byte {
	switch v := v.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	default:
		return bin.Encode(v)
	}
}

func (t *Tree) Put(key, value []byte) (err error) {
	t.puts = map[string]*node{}
	defer func() {
		t.puts = nil
	}()

	root, err := t.put(make([]byte, 0, 10), key, value)
	if err != nil {
		return
	}

	// save to db
	for path, nd := range t.puts {
		err = t.db.Put([]byte(path), nd.encode())
		if err != nil {
			return
		}
	}

	// success
	t.root = root
	return
}

func (t *Tree) Get(key []byte) (value []byte, err error) {
	value, _, err = t.proof(make([]byte, 0, 10), key)
	return
}

func (t *Tree) GetVar(key, v interface{}) (err error) {
	data, err := t.Get(encode(key))
	if err != nil {
		return
	}
	return decode(data, v)
}

func (t *Tree) GetProof(key []byte) (value, proof, root []byte, err error) {
	if value, proof, err = t.proof(make([]byte, 0, 10), key); err != nil {
		return
	}
	root, err = t.Root()
	return
}

func (t *Tree) AppendingProof(newKey, value []byte) (proof, root []byte, err error) {
	t.puts = map[string]*node{}
	defer func() {
		t.puts = nil
	}()
	buf := make([]byte, 0, 10)
	root, err = t.put(buf, newKey, value)
	if err != nil {
		return
	}
	_, proof, err = t.proof(buf, newKey)
	if err != nil {
		return
	}
	return
}

//--------------------------------------------------
func (t *Tree) getNode(key []byte) (nd *node, err error) {
	if t.puts != nil {
		if nd, ok := t.puts[string(key)]; ok {
			return nd, nil
		}
	}
	data, err := t.db.Get(key)
	if err != nil {
		return
	}
	if len(data) == 0 {
		return nil, nil
	}
	nd = new(node)
	err = nd.decode(data)
	return
}

func idx(key []byte, lv int) uint8 {
	if lv%2 == 0 {
		return key[lv/2] >> 4
	} else {
		return key[lv/2] & 0x0f
	}
}

func (t *Tree) put(path, key, value []byte) (newHash []byte, err error) {
	nd, err := t.getNode(path)
	if err != nil {
		return
	}
	if nd == nil {
		nd = &node{key: key, value: value}

	} else if bytes.Equal(nd.key, key) {
		nd.value = value

	} else {
		lv := len(path)
		if nd.key != nil {
			k, v := nd.key, nd.value
			i := idx(k, lv)
			nd.key, nd.value = nil, nil
			nd.hashes[i], err = t.put(append(path, i), k, v)
			if err != nil {
				return
			}
		}
		i := idx(key, lv)
		nd.hashes[i], err = t.put(append(path, i), key, value)
		if err != nil {
			return
		}
	}
	t.puts[string(path)] = nd
	return nd.hash(), err
}

func (t *Tree) proof(path, key []byte) (value, proof []byte, err error) {
	nd, err := t.getNode(path)
	if err != nil {
		return
	}
	if nd == nil { // null leaf
		return nil, nil, errKeyNotFound
	}
	if nd.key != nil { // leaf
		if !bytes.Equal(nd.key, key) {
			// todo: return proof of empty
			return nil, nil, errKeyNotFound
		}
		return nd.value, nil, nil
	}
	i := idx(key, len(path))
	value, proof, err = t.proof(append(path, i), key)
	proof = append(proof, nd.proof(int(i))...)
	return
}
