package patricia

type MemoryStorage map[string][]byte

func (m MemoryStorage) Get(k []byte) ([]byte, error) {
	return m[string(k)], nil
}

func (m MemoryStorage) Put(key, val []byte) error {
	m[string(key)] = val
	return nil
}
